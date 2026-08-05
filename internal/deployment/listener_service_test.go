package deployment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	config "github.com/salmanrf/capybara-cloud/pkg/utils"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
	"github.com/salmanrf/capybara-cloud/shared/utils"
)

func TestDeployListener(t *testing.T) {
	in_chan := make(chan DeployRequest)
	out_chan := make(chan DeployStepResult)

	deployment_service := StubService{}
	masbro_service := StubMasbroService{}
	listener_service := NewListener(in_chan, out_chan, &deployment_service, &masbro_service)

	mock_app := database.Application{
		Name: "testapp",
		Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
	}
	mock_dp := database.ApplicationDeployment{
		AppID: mock_app.AppID,
		Status: shared_deployment.DEPLOY_STATUS_INITIATED,
		ArtifactsPath: "",
		BuildPath: "",
		VersionNumber: 1,
	}
	mock_app_cfg := database.ApplicationConfig{
		Port: 3000,
	}

	deploy_request := DeployRequest{
		ApplicationDto: mock_app,
		DeploymentDto: mock_dp,
		ApplicationConfig: mock_app_cfg,
	}

	t.Run("should call build extract when status is DEPLOY_STATUS_INITIATED = 1", func (t *testing.T) {
		defer func () {
			deployment_service.Clear()
		}()
		
		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_INITIATED
		
		go listener_service.Listen()
		in_chan <- req 

		timer := time.NewTimer(time.Second)

		select {
		case <- timer.C:
			t.Fatal("unexpected timeout")
		case <- out_chan:
			want_extract_called := 1
			got_extract_called := deployment_service.extract_n_calls
	
			if got_extract_called != want_extract_called {
				t.Errorf("got deployment_service.Extract called %d time, want 1", got_extract_called)
			}
		}
	})

	t.Run("should call build when status is DEPLOY_STATUS_BUILD_EXTRACTED", func (t *testing.T) {
		defer func () {
			deployment_service.Clear()
		}()
		
		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED
		
		go listener_service.Listen()
		in_chan <- req

		timer := time.NewTimer(time.Second)

		select {
		case <- timer.C:
			t.Fatal("unexpected timeout")
		case <- out_chan:
			want_extract_called := 0
			got_extract_called := deployment_service.extract_n_calls

			if got_extract_called != want_extract_called {
				t.Errorf("got extract called %d time, want %d", got_extract_called, want_extract_called)
			}

			want_build_called := 1
			got_build_called := deployment_service.build_n_calls

			if got_build_called != want_build_called {
				t.Errorf("got build called %d time, want %d", got_build_called, want_build_called)
			}
		}
	})

	t.Run("should call push when status is DEPLOY_STATUS_BUILD_IMAGE_BUILT", func (t *testing.T) {
		defer func () {
			deployment_service.Clear()
		}()
		
		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		
		go listener_service.Listen()
		in_chan <- req

		timer := time.NewTimer(time.Second)

		select {
		case <- timer.C:
			t.Fatal("unexpected timeout")
		case <- out_chan:
			want_extract_called := 0
			got_extract_called := deployment_service.extract_n_calls
			if got_extract_called != want_extract_called {
				t.Errorf("got extract called %d time, want %d", got_extract_called, want_extract_called)
			}

			want_build_called := 0
			got_build_called := deployment_service.build_n_calls
			if got_build_called != want_build_called {
				t.Errorf("got build called %d time, want %d", got_build_called, want_build_called)
			}

			want_push_called := 1
			got_push_called := deployment_service.push_n_calls

			if got_push_called != want_push_called {
				t.Errorf("got push called %d time, want %d", got_push_called, want_push_called)
			}
		}
	})

	t.Run("should call start when status is DEPLOY_STATUS_BUILD_IMAGE_PUSHED", func (t *testing.T) {
		defer func () {
			deployment_service.Clear()
		}()

		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED

		go listener_service.Listen()
		in_chan <- req

		timer := time.NewTimer(time.Second)
		<- timer.C

		want_push_called := 0
		got_push_called := deployment_service.push_n_calls

		if got_push_called != want_push_called {
			t.Errorf("got push called %d time, want %d", got_push_called, want_push_called)
		}

		want_start_called := 1
		got_start_called := masbro_service.start_n_calls

		if got_start_called != want_start_called {
			t.Errorf("got masbro start called %d time, want %d", got_start_called, want_start_called)
		}
	})
}

func TestDeployExtract(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		&StubDocker{},
		app_service,
		port_service,
		deployment_repository,
		make(chan DeployRequest, 1),
	)

	t.Run("should perform artifact extraction when status is DEPLOY_STATUS_INITIATED = 1", func (t *testing.T) {
		cfg := config.GetConfig()
		cfg.BASE_BUILD_PATH = "/tmp/tests/build-artifacts"
		config.SetConfig(cfg)
		
		// * Uses /deployment/test-samples/<filename>
		type test struct{
			tid   							uuid.UUID
			name 								string
			artifact_filename 	string
			artifact_path 			string
		}
		tests := []test{
			{
				uuid.New(),
				"abcd",
				"abcd.tar.gz",
				"",
			},
			{
				uuid.New(),
				"app-html",
				"sample-html.tar.gz",
				"",
			},
			{
				uuid.New(),
				"app-vue",
				"sample-vue.tar.gz",
				"",
			},
		}

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// * For inspecting the content of a tarball
		// * Use output from tar to compare with ReadDir 
		/* Sample output:
			./
			./d.txt
			./b.txt
			./a.txt
			./c.txt
		*/
		tar, err := exec.LookPath("tar")
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}


		// * Cleanup build directories
		defer func () {
			os.Remove(cfg.BASE_BUILD_PATH)
		}()

		for _, tt := range tests {
			t.Run("should match file count by tar -tzf after extraction", func (t *testing.T) {
				tt.artifact_path = path.Join(pwd, "test-samples", tt.artifact_filename)
				
				mock_app := database.Application{
					Name: tt.name,
					Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
				}	

				want_build_path := getBuildDirPath(cfg, "localfs", mock_app.Name)
				want_file_count := 0

				// * Store output of tar -tzf
				output := bytes.NewBuffer([]byte{})
				cmd := exec.Command(tar, "-tzf", tt.artifact_path)
				cmd.Stdout = output
				if err := cmd.Run(); err != nil {
					t.Fatalf("got error checking actual tar content %v, want nil", err)
				}
				// * Clean trailing newline
				output_str := strings.Trim(output.String(), " \n")
				
				// * Format tar output, clean '.' and take only the first entry after the first slash
				// * eg, ./abcd.txt -> abcd.txt
				// * eg, ./a/b -> a
				// * Track occurence in map
				entries := strings.Split(output_str, "\n")
				file_entries := make([]string, 0)
				base_map := make(map[string]bool)
				for _, ent := range entries {
					dot_cleaned := strings.Trim(ent, ".")
					if dot_cleaned == "" {
						continue
					}
					
					parts :=  strings.Split(dot_cleaned, "/")
					if len(parts) < 2 {
						continue
					}

					clean := parts[1]
					if clean == "" {
						continue
					}
					if base_map[clean] {
						continue
					} else {
						base_map[clean] = true	
					}

					file_entries = append(file_entries, clean)
				}
				want_file_count = len(file_entries)

				mock_dp := database.ApplicationDeployment{
					AppID: mock_app.AppID,
					Status: shared_deployment.DEPLOY_STATUS_INITIATED,
					ArtifactsPath: tt.artifact_path,
					BuildPath: want_build_path,
					VersionNumber: 1,
				}
				mock_app_cfg := database.ApplicationConfig{
					Port: 3000,
				}

				deploy_request := DeployRequest{
					ApplicationDto: mock_app,
					DeploymentDto: mock_dp,
					ApplicationConfig: mock_app_cfg,
				}

				res, err := deployment_service.Extract(deploy_request)
				if err != nil {
					t.Errorf("got deployment build step error %v, want nil", err)
				}

				dep := res.DeploymentDto
				got_new_status := dep.Status
				want_new_status := shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED

				if got_new_status != int32(want_new_status) {
					t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
				}

				got_entries, err := os.ReadDir(want_build_path)
				if err != nil {
					t.Errorf("got read build dir %s error %v, want nil", want_build_path, err)
				}

				got_file_count := len(got_entries)
				if got_file_count != want_file_count {
					t.Errorf("got %d file count after extract, want %d", got_file_count, want_file_count)
				}
			})
		}
	})
} 

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
				deploy_request.DeploymentDto = *res.DeploymentDto

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

				got_dp := res.DeploymentDto
				if got_dp == nil {
					t.Fatalf("got deployment result %v, want non-nil", got_dp)
				}

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

func TestDeployPush(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	cfg := config.GetConfig()
	cfg.DOCKER_REGISTRY = "salmanrf"
	cfg.BASE_BUILD_PATH = "/tmp/masmasbro/builds"
	cfg.BASE_ARTIFACT_PATH = "/tmp/masmasbro/artifacts"
	config.SetConfig(cfg)

	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	docker := &StubDocker{}
	deployment_service := NewService(
		ctx,
		docker,
		app_service,
		port_service,
		deployment_repository,
		make(chan DeployRequest, 1),
	)

	t.Run("should return error if container image not found", func (t *testing.T) {
		docker.find_one_image_by_name_return = nil
		docker.find_one_image_by_name_error = nil

		deploy_req := DeployRequest{
			ApplicationDto: database.Application{},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: database.ApplicationDeployment{},
		}

		res, err := deployment_service.Push(deploy_req)
		if err == nil {
			t.Fatalf("got error '%v', want '%v'", err, errors.New("image_not_found"))
		}

		got_dp := res.DeploymentDto
		if got_dp == nil {
			t.Errorf("got new deployment nil, want non-nil")
		}

		got_new_status := got_dp.Status
		want_new_status := -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		if got_new_status != int32(want_new_status) {
			t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
		}
	})

	t.Run("should return error if push fails", func (t *testing.T) {
		mock_image_summary := &image.Summary{
			ID: "abcd",
			RepoTags: []string{"mrfreshgallery-123"},
		}
		docker.find_one_image_by_name_return = mock_image_summary
		docker.find_one_image_by_name_error = nil
		docker.push_return = errors.New("fatal error")

		defer docker.Clear()

		mock_dp := database.ApplicationDeployment{
			ContainerImgName: "mrfreshgallery-backend-123",
		}
		deploy_req := DeployRequest{
			ApplicationDto: database.Application{
				Name: "mrfreshgallery",
			},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: mock_dp,
		}

		res, got_err := deployment_service.Push(deploy_req)
		want_err_msg := fmt.Sprintf("image_push_failed: %s", docker.push_return.Error())
		want_error := errors.New(want_err_msg)
		if got_err == nil {
			t.Fatalf("got error '%v', want '%v'", got_err, want_error)
		}

		got_dp := res.DeploymentDto
		if got_dp == nil {
			t.Errorf("got new deployment nil, want non-nil")
		}

		got_new_status := got_dp.Status
		want_new_status := -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		if got_new_status != int32(want_new_status) {
			t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
		}
	})
	
	t.Run("should correctly uses internal Docker API", func (t *testing.T) {
		cfg := config.GetConfig()
		cfg.DOCKER_REGISTRY = "docker.io/test"
		config.SetConfig(cfg)
		
		mock_image_summary := &image.Summary{
			ID: "abcd",
			RepoTags: []string{"mrfreshgallery-backend-123"},
		}
		docker.find_one_image_by_name_return = mock_image_summary
		docker.find_one_image_by_name_error = nil

		defer docker.Clear()

		mock_dp := database.ApplicationDeployment{
			ContainerImgName: 
				fmt.Sprintf("%s/mrfreshgallery-backend-123", cfg.DOCKER_REGISTRY),
		}
		deploy_req := DeployRequest{
			ApplicationDto: database.Application{
				Name: "mrfreshgallery",
			},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: mock_dp,
		}

		_, err := deployment_service.Push(deploy_req)
	
		parts := strings.Split(mock_dp.ContainerImgName, fmt.Sprintf("%s/", cfg.DOCKER_REGISTRY))
		want_repo_tag := parts[1]

		got_find_called_n_times := docker.find_one_image_by_name_return_n_calls
		want_find_called_n_times := 1
		if got_find_called_n_times != want_find_called_n_times {
			t.Errorf("got find one image called %d times, want %d times", got_find_called_n_times, want_find_called_n_times)
		}

		got_find_called_with_name := docker.find_one_image_by_name_return_call_args[0]
		want_find_called_with_name := want_repo_tag
		if got_find_called_with_name != want_find_called_with_name {
			t.Errorf("got find one image called with '%s', want '%s'", got_find_called_with_name, want_find_called_with_name)
		}

		got_push_called_n_times := docker.push_return_n_calls
		want_push_called_n_times := 1
		if got_push_called_n_times != want_push_called_n_times {
			t.Errorf("got push called %d times, want %d times", got_push_called_n_times, want_push_called_n_times)
		}
		
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_push_called_with_name := docker.push_return_call_args[0]
		want_push_called_with_name := want_repo_tag
		if got_push_called_with_name!= want_push_called_with_name {
			t.Errorf("got push called with container image name %s, want %s", got_push_called_with_name, want_push_called_with_name)
		}
	})

	t.Run("should perform push when image exists", func (t *testing.T) {
		mock_image_summary := &image.Summary{
			ID: "abcd",
			RepoTags: []string{"mrfreshgallery-123"},
		}
		docker.find_one_image_by_name_return = mock_image_summary
		docker.find_one_image_by_name_error = nil

		defer docker.Clear()

		deploy_req := DeployRequest{
			ApplicationDto: database.Application{
				Name: "mrfreshgallery",
			},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: database.ApplicationDeployment{},
		}

		res, err := deployment_service.Push(deploy_req)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_dp := res.DeploymentDto
		if got_dp == nil {
			t.Error("got updated deployment nil, want non-nil")
		}

		got_new_status := got_dp.Status
		want_new_status := shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		if got_new_status != int32(want_new_status) {
			t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
		}
	})
}