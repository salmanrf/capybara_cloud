package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmanrf/capybara-cloud/api"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/deployment"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
	config "github.com/salmanrf/capybara-cloud/pkg/utils"
	"github.com/salmanrf/capybara-cloud/shared/docker"
)

func create_db_conn(ctx context.Context, db_uri string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(ctx, db_uri)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return dbpool
}

func setup() (context.Context, config.Config, *pgxpool.Pool, error) {
	pwd, _ := os.Getwd()
	envpath := filepath.Join(pwd, ".env")

	cfg, err := config.LoadConfig(envpath)
	if err != nil {
		return nil, cfg, nil, err
	}

	ctx := context.Background()
	dbpool := create_db_conn(ctx, cfg.POSTGRES_URI)

	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, cfg, nil, fmt.Errorf("unable to ping database: %w", err)
	}

	fmt.Println("Database connection established")

	return ctx, cfg, dbpool, nil
}

func main() {
	ctx, cfg, db_conn, err := setup()
	if err != nil {
		log.Fatal(err)
	}
	defer db_conn.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Server encountered a panic", err)
		}

		fmt.Println("Server is stopped")
	}()

	queries := database.New(db_conn)
	application_repository := application.NewRepository(ctx, queries)
	deployment_repository := deployment.NewDeploymentRepository(ctx, queries)
	deploy_chan := make(chan deployment.DeployRequest)
	
	user_service := user.NewService(ctx, queries)
	auth_service := auth.NewService(ctx, user_service)
	org_service := organization.NewService(ctx, db_conn, queries, user_service)
	project_service := project.NewService(ctx, db_conn, queries, user_service)
	application_service := application.NewService(ctx, application_repository, project_service)
	port_allocator_service := deployment.NewPortAllocatorService()
	docker_service, err := docker.New(
		docker.DockerConfig{
			Registry: cfg.DOCKER_REGISTRY,
			AccessToken: cfg.DOCKER_ACCESS_TOKEN,
			Username: cfg.DOCKER_USER,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	deployment_service := deployment.NewService(ctx, docker_service, application_service, port_allocator_service, deployment_repository, deploy_chan)
	jwt_utils := auth_utils.NewJWTUtils(cfg.AUTH_JWT_SECRET)

	api_server := api.NewAPIServer(
		ctx,
		application_service,
		deployment_service,
		user_service,
		auth_service,
		org_service,
		project_service,
		jwt_utils,
	)

	address := fmt.Sprintf(":%s", cfg.API_PORT)
	fmt.Printf("Starting API server on %s\n", address)
	if err := http.ListenAndServe(address, api_server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}