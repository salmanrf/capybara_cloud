package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmanrf/capybara-cloud/apps/backend/api"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/application"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/auth"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/deployment"
	masbro_worker "github.com/salmanrf/capybara-cloud/apps/backend/internal/masbro-worker"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/organization"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/project"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/user"
	locutils "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/docker"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/logger"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

func create_db_conn(ctx context.Context, db_uri string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(ctx, db_uri)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return dbpool
}

type futils struct {}
func (_ *futils) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func setup() (func (), context.Context, locutils.Config, *slog.Logger, *pgxpool.Pool, error) {
	futil := &futils{}

	pwd, _ := os.Getwd()
	envpath := path.Join(pwd, ".env")

	cfg, err := locutils.LoadConfig(envpath)
	if err != nil {
		return nil, nil, cfg, nil, nil, err
	}
	
	logpath := path.Join(pwd, "apps.backend.logs")
	logger, log_cleanup, err := logger.InitLogger(logpath, futil)
	if err != nil {
		return nil, nil, cfg, nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbpool := create_db_conn(ctx, cfg.POSTGRES_URI)

	err = dbpool.Ping(ctx)
	if err != nil {
		cancel()
		return nil, nil, cfg, nil, nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Debug("Database connection established")

	teardown := func () {
		log_cleanup()
		cancel()
	}

	return teardown, ctx, cfg, logger, dbpool, nil
}

func main() {
	teardown, ctx, cfg, logger, db_conn, err := setup()
	if err != nil {
		log.Fatal(err)
	}
	defer db_conn.Close()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("Server encountered a panic", "error", err)
		}
		logger.Debug("Server is stopped")
		teardown()
	}()

	queries := database.New(db_conn)
	application_repository := application.NewRepository(ctx, queries)
	deployment_repository := deployment.NewDeploymentRepository(ctx, queries)
	deploy_in_chan := make(chan deployment.DeployRequest)
	deploy_out_chan := make(chan deployment.DeployStepResult)
	
	user_service := user.NewService(ctx, logger, queries)
	auth_service := auth.NewService(ctx, logger, user_service)
	org_service := organization.NewService(ctx, logger, db_conn, queries, user_service)
	project_service := project.NewService(ctx, logger, db_conn, queries, user_service)
	application_service := application.NewService(ctx, logger, application_repository, project_service)
	port_allocator_service := shared_deployment.NewPortAllocatorService()
	docker_service, err := docker.New(
		docker.DockerConfig{
			Registry: cfg.DOCKER_REGISTRY,
			Namespace: cfg.DOCKER_NAMESPACE,
			AccessToken: cfg.DOCKER_ACCESS_TOKEN,
			Username: cfg.DOCKER_USER,
		},
		port_allocator_service,
	)
	if err != nil {
		logger.Error("Unable to create docker client", "error", err)
		teardown()
		os.Exit(1)
	}
	masbro_service := masbro_worker.New(docker_service)
	deployment_service := deployment.NewService(
		ctx, 
		logger,
		docker_service, 
		application_service, 
		masbro_service,
		port_allocator_service, 
		deployment_repository, 
		deploy_in_chan,
	)
	listener_service := deployment.NewListener(
		ctx,
		logger,
		deploy_in_chan,
		deploy_out_chan,
		deployment_service,
		masbro_service,
	)
	jwt_utils := utils.NewJWTUtils(cfg.AUTH_JWT_SECRET, cfg.AUTH_JWT_ISSUER, []string{cfg.AUTH_JWT_AUDIENCE})

	api_server := api.NewAPIServer(
		logger,
		application_service,
		deployment_service,
		user_service,
		auth_service,
		org_service,
		project_service,
		jwt_utils,
	)

	address := fmt.Sprintf(":%s", cfg.API_PORT)
	logger.Debug("Starting listeners")
	go listener_service.Listen()
	logger.Debug("Starting API server", "address", address)
	if err := http.ListenAndServe(address, api_server); err != nil {
		logger.Error("Server failed", "error", err, "address", address)
		teardown()
		os.Exit(1)
	}
}