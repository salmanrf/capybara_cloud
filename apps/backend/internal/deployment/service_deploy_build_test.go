package deployment

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/application"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"

	config "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
)

func TestDeployBuild(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}
	stub_docker := &StubDocker{}

	cfg := config.GetConfig()
	cfg.DOCKER_REGISTRY = "docker.io"
	cfg.DOCKER_NAMESPACE = "salmanrf"
	cfg.BASE_BUILD_PATH = "/tmp/masmasbro/builds"
	cfg.BASE_ARTIFACT_PATH = "/tmp/masmasbro/artifacts"
	config.SetConfig(cfg)

	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		stub_docker,
		app_service,
		&StubMasbroService{},
		port_service,
		deployment_repository,
		make(chan DeployRequest, 1),
	)

	// * Uses /deployment/test-samples/<filename>
	type test struct{
		tid   							uuid.UUID
		name 								string
		types								string
		artifact_filename 	string
		artifact_path 			string
	}

	tests := []test{
		{
			uuid.New(),
			"abcd",
			shared_deployment.APP_TYPE_NODEJS_CONTAINER,
			"abcd.tar.gz",
			"",
		},
		{
			uuid.New(),
			"app-html",
			shared_deployment.APP_TYPE_NODEJS_CONTAINER,
			"sample-html.tar.gz",
			"",
		},
		{
			uuid.New(),
			"app-vue",
			shared_deployment.APP_TYPE_NODEJS_CONTAINER,
			"sample-vue.tar.gz",
			"",
		},
	}

	t.Run("should return error if deployment's build directory doesn't exist", func (t *testing.T) {
		for _, tt := range tests {
			want_build_path := getBuildDirPath(cfg, "localfs", tt.name)
			t.Run(fmt.Sprintf("should return error if %s doesn't exist", want_build_path), func (t *testing.T) {
				mock_app := database.Application{
					Name: tt.name,
					Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
				}	

				created_at := pgtype.Timestamp{}
				created_at.Scan(time.Now())
				mock_dp := database.ApplicationDeployment{
					AppID: mock_app.AppID,
					Status: shared_deployment.DEPLOY_STATUS_INITIATED,
					ArtifactsPath: tt.artifact_path,
					BuildPath: want_build_path,
					VersionNumber: 1,
					CreatedAt: created_at,
				}
				mock_app_cfg := database.ApplicationConfig{
					Port: pgtype.Int4{Int32: 3000, Valid: true},
				}

				deploy_request := DeployRequest{
					ApplicationDto: mock_app,
					DeploymentDto: mock_dp,
					ApplicationConfig: mock_app_cfg,
				}
				
				os.Remove(want_build_path)

				res, err := deployment_service.Build(deploy_request)
				if err == nil {
					t.Fatalf("got error nil, want error")
				}

				got_new_status := res.DeploymentDto.Status
				want_new_status := int(mock_dp.Status)
				
				if int(got_new_status) != want_new_status {
					t.Errorf("got new status %d, want %d", got_new_status, want_new_status)
				} 
			})
		}
	})

	t.Run("should return error if the proper docker template does not exist in DOCKER_TEMPLATES_DIR", func (t *testing.T) {
		// ? Use only one case
		for _, tt := range tests[:1] {
			cfg := config.GetConfig()
			cfg.DOCKER_TEMPLATES_DIR = t.TempDir()
			config.SetConfig(cfg)
			defer func () {
				cfg.DOCKER_TEMPLATES_DIR = ""
				config.SetConfig(cfg)
			}()
			
			mock_app := database.Application{
				Name: tt.name,
				Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
			}	
				
			want_build_path := getBuildDirPath(cfg, "localfs", mock_app.Name)

			created_at := pgtype.Timestamp{}
			created_at.Scan(time.Now())
			mock_dp := database.ApplicationDeployment{
				AppID: mock_app.AppID,
				Status: shared_deployment.DEPLOY_STATUS_INITIATED,
				ArtifactsPath: tt.artifact_path,
				BuildPath: want_build_path,
				VersionNumber: 1,
				CreatedAt: created_at,
			}
			mock_app_cfg := database.ApplicationConfig{
				Port: pgtype.Int4{Int32: 3000, Valid: true},
			}

			deploy_request := DeployRequest{
				ApplicationDto: mock_app,
				DeploymentDto: mock_dp,
				ApplicationConfig: mock_app_cfg,
			}

			err := utils.EnsureDirExists(want_build_path)
			if err != nil {
				t.Fatal("got unexpected error ensuring build path", err)
			}

			_, err = deployment_service.Build(deploy_request)
			if err == nil {
				t.Fatal("got error nil, want error")
			}

			want_err_msg_containing := "Unable to locate template Dockerfile"
			if msg := err.Error(); !strings.Contains(msg, want_err_msg_containing) {
				t.Fatalf("got error message %s, want containing %s", msg, want_err_msg_containing)
			}
		}
	})

	t.Run("should perform artifact build when status is DEPLOY_STATUS_BUILD_EXTRACTED", func (t *testing.T) {
		cfg := config.GetConfig()
		cfg.BASE_BUILD_PATH = "/tmp/tests/build-artifacts"

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		cfg.DOCKER_TEMPLATES_DIR = path.Join(pwd, "templates")
		config.SetConfig(cfg)

		defer func () {
			os.Remove(cfg.BASE_BUILD_PATH)
		}()

		for _, tt := range tests[1:] {
			tt.artifact_path = path.Join(pwd, "test-samples", tt.artifact_filename)
				
			mock_app := database.Application{
				Name: tt.name,
				Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
			}	

			want_build_path := getBuildDirPath(cfg, "localfs", mock_app.Name)
			created_at := pgtype.Timestamp{}
			created_at.Scan(time.Now())
			mock_dp := database.ApplicationDeployment{
				AppID: mock_app.AppID,
				Status: shared_deployment.DEPLOY_STATUS_INITIATED,
				ArtifactsPath: tt.artifact_path,
				BuildPath: want_build_path,
				VersionNumber: 1,
				CreatedAt: created_at,
			}
			mock_app_cfg := database.ApplicationConfig{
				Port: pgtype.Int4{Int32: 3000, Valid: true},
			}

			deploy_request := DeployRequest{
				ApplicationDto: mock_app,
				DeploymentDto: mock_dp,
				ApplicationConfig: mock_app_cfg,
			}
			app_name_slug := utils.Slugify(mock_app.Name)
			want_registry := cfg.DOCKER_REGISTRY
			want_namespace := cfg.DOCKER_NAMESPACE
			want_repository := app_name_slug
			want_tag := getContainerTag(mock_dp)
			want_full_ref := fmt.Sprintf(
				"%s/%s/%s:%s",
				want_registry,
				want_namespace,
				want_repository,
				want_tag,
			)

			t.Run(fmt.Sprintf("should build container image '%s'", want_full_ref), func (t *testing.T) {
				// * Perform extract first
				res, err := deployment_service.Extract(deploy_request)
				if err != nil {
					t.Fatal(err)
				}
				deploy_request.DeploymentDto = res.DeploymentDto

				stub_docker.Clear()
				res, err = deployment_service.Build(deploy_request)
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}

				if stub_docker.build_n_calls != 1 {
					t.Fatalf("got docker build called %d times, want 1", stub_docker.build_n_calls)
				}
				got_build_args := stub_docker.build_call_args[0]
				if got_build_args.Tag != want_full_ref {
					t.Errorf("got docker build tag %s, want %s", got_build_args.Tag, want_full_ref)
				}
				if got_build_args.ContextPath != want_build_path {
					t.Errorf("got docker build context %s, want %s", got_build_args.ContextPath, want_build_path)
				}

				got_new_status := res.DeploymentDto.Status
				want_new_status := int(mock_dp.Status)
				if got_new_status != int32(want_new_status) {
					t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
				}
			})
		}
	})

	t.Run("should remove build directory once finished", func (t *testing.T) {
		for _, tt := range tests {
			mock_app := database.Application{
				Name: tt.name,
				Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
			}	

			want_build_path := getBuildDirPath(cfg, "localfs", mock_app.Name)

			created_at := pgtype.Timestamp{}
			created_at.Scan(time.Now())
			mock_dp := database.ApplicationDeployment{
				AppID: mock_app.AppID,
				Status: shared_deployment.DEPLOY_STATUS_INITIATED,
				ArtifactsPath: tt.artifact_path,
				BuildPath: want_build_path,
				VersionNumber: 1,
				CreatedAt: created_at,
			}
			mock_app_cfg := database.ApplicationConfig{
				Port: pgtype.Int4{Int32: 3000, Valid: true},
			}

			deploy_request := DeployRequest{
				ApplicationDto: mock_app,
				DeploymentDto: mock_dp,
				ApplicationConfig: mock_app_cfg,
			}

			err := utils.EnsureDirExists(want_build_path)
			if err != nil {
				t.Fatal("got unexpected error ensuring build path", err)
			}

			_, err = deployment_service.Build(deploy_request)
			if err != nil {
				t.Fatal("got unexpected error, want nil", err)
			}

			_, err = os.ReadDir(want_build_path)
			if err == nil {
				t.Fatal("got error nil, want error no such file / directory does")
			}
		}
	})
}