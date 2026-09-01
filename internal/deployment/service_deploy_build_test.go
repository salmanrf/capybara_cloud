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
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/shared/utils"

	config "github.com/salmanrf/capybara-cloud/pkg/utils"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
)

func TestDeployBuild(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	cfg := config.GetConfig()
	cfg.DOCKER_REGISTRY = "salmanrf"
	cfg.BASE_BUILD_PATH = "/tmp/masmasbro/builds"
	cfg.BASE_ARTIFACT_PATH = "/tmp/masmasbro/artifacts"
	config.SetConfig(cfg)

	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		&StubDocker{},
		app_service,
		&StubMasbroService{},
		port_service,
		deployment_repository,
		make(chan DeployRequest, 1),
	)

	docker, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatal("unable to instantiate docker client", err)
	}

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
					Port: 3000,
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
				want_new_status := -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
				
				if int(got_new_status) != want_new_status {
					t.Errorf("got new status %d, want %d", got_new_status, want_new_status)
				} 
			})
		}
	})

	t.Run("should return error if the proper docker template does not exist, or copy fails", func (t *testing.T) {
		// ? Run only one test
		for _, tt := range tests[:1] {
			wd, _ := os.Getwd()
			template_path_old := path.Join(wd, "templates", "Dockerfile")
			template_path_new := fmt.Sprintf("%s_new", template_path_old)

			err := os.Rename(template_path_old, template_path_new)
			if err != nil {
				t.Fatal("got unexpected error", err)
			}

			defer func () {
				temp := template_path_old
				template_path_old = template_path_new
				template_path_new = temp

				err := os.Rename(template_path_old, template_path_new)
				if err != nil {
					t.Fatal("got unexpected error", err)
				}
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
				Port: 3000,
			}

			deploy_request := DeployRequest{
				ApplicationDto: mock_app,
				DeploymentDto: mock_dp,
				ApplicationConfig: mock_app_cfg,
			}

			err = utils.EnsureDirExists(want_build_path)
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
		config.SetConfig(cfg)

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

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
				Port: 3000,
			}

			deploy_request := DeployRequest{
				ApplicationDto: mock_app,
				DeploymentDto: mock_dp,
				ApplicationConfig: mock_app_cfg,
			}
			want_image_name := getContainerImageName(deploy_request.ApplicationDto, deploy_request.DeploymentDto)
			
			t.Run(fmt.Sprintf("should build container image %s", want_image_name), func (t *testing.T) {
				// * Perform extract first
				res, err := deployment_service.Extract(deploy_request)
				if err != nil {
					t.Fatal(err)
				}
				deploy_request.DeploymentDto = res.DeploymentDto

				res, err = deployment_service.Build(deploy_request)
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				
				image_list, err := docker.ImageList(
					ctx, 
					client.ImageListOptions{}, 
				)
				if err != nil {
					t.Fatal("got error from docker image list", err)
				}

				got_found := false
				want_found := true
				var found image.Summary 
				for _, ct := range image_list.Items {
					for _, n := range ct.RepoTags {
						if n != "" && strings.Contains(want_image_name, n) {
							got_found = true
							found = ct
							break
						}
						if got_found {
							break
						}
					}
				}			

				if got_found != want_found {
					t.Fatalf("got container '%s' found == %v, want %v", want_image_name, got_found, want_found)
				}

				defer func (ct image.Summary) {
					if ct.ID != "" {
						_, err = docker.ImageRemove(ctx, found.ID, client.ImageRemoveOptions{Force: true})
						if err != nil {
							t.Logf("got unexpected error from delete image %v", err)
						}
					}
				}(found)


				got_new_status := res.DeploymentDto.Status
				want_new_status := shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
				if got_new_status != int32(want_new_status) {
					t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
				}

				got_image_name := res.DeploymentDto.ContainerImgName
				if got_image_name != want_image_name {
					t.Errorf("got deployment image name %s, want %s", got_image_name, want_image_name)
				}

				got_docker_registry := res.DeploymentDto.ContainerRegistry
				want_docker_registry := cfg.DOCKER_REGISTRY
				if got_docker_registry != want_docker_registry {
					t.Errorf("got docker registry %v, want %v", got_docker_registry, want_docker_registry)
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
				Port: 3000,
			}

			deploy_request := DeployRequest{
				ApplicationDto: mock_app,
				DeploymentDto: mock_dp,
				ApplicationConfig: mock_app_cfg,
			}

			err = utils.EnsureDirExists(want_build_path)
			if err != nil {
				t.Fatal("got unexpected error ensuring build path", err)
			}

			_, err = deployment_service.Build(deploy_request)
			if err != nil {
				t.Fatal("got unexpected error, want nil", err)
			}

			_, err := os.ReadDir(want_build_path)
			if err == nil {
				t.Fatal("got error nil, want error no such file / directory does")
			}
		}
	})
}