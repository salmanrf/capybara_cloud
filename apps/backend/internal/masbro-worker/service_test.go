package masbro_worker

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
	docker "github.com/salmanrf/capybara-cloud/packages/shared-go/docker"
)

func TestStart(t *testing.T) {
	stub_docker := &StubDocker{}
	masbro_service := New(stub_docker)

	mock_app := database.Application{
		Name: "testapp",
		Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
	}
	mock_dp := database.ApplicationDeployment{
		AppID: mock_app.AppID,
		Status: shared_deployment.DEPLOY_STATUS_INITIATED,
		ContainerImgName: "docker.io/capybaracloud/testapp:latest",
		ArtifactsPath: "",
		BuildPath: "",
		VersionNumber: 1,
	}
	mock_app_cfg := database.ApplicationConfig{
		Port: 3000,
		VariablesJson: []byte(
			`
				{
					"CLIENT_ID": "abcd",
					"CLIENT_SECRET": "zxcvbnm"
				}
			`,
		),
	}

	deploy_request := shared_deployment.DeployRequest{
		ApplicationDto: mock_app,
		DeploymentDto: mock_dp,
		ApplicationConfig: mock_app_cfg,
	}
	
	t.Run("should return error if variables is not a valid json", func (t * testing.T) {
		defer func () {
			stub_docker.Clear()
		}()	

		stub_docker.pull_error = nil

		mock_instance := database.DeploymentInstance{
			ContainerName: "testapp-v1-abcd",
			ContainerPort: mock_app_cfg.Port,
		}
		mock_app_cfg := mock_app_cfg
		mock_app_cfg.VariablesJson = []byte("invalid json")
		req := deploy_request
		req.ApplicationConfig = mock_app_cfg

		_, start_err := masbro_service.Start(req, mock_instance)
		if start_err == nil {
			t.Fatalf("got error nil, want '%v'", errors.New("invalid_json"))
		}
	})

	t.Run("should run the container with the correct configuration", func (t * testing.T) {
		defer func () {
			stub_docker.Clear()
		}()	

		stub_docker.pull_error = nil
		stub_docker.run_return = &docker.DockerRunResult{
			HostPort: 9999,
		}

		mock_instance := database.DeploymentInstance{
			ContainerName: "testapp-v1-abcd",
			ContainerPort: mock_app_cfg.Port,
		}

		start_res, start_err := masbro_service.Start(deploy_request, mock_instance)
		if start_err != nil {
			t.Fatalf("got unexpected error '%v', want nil", start_err)
		}

		got_run_called := stub_docker.run_n_calls
		want_run_called := 1
		if got_run_called != want_run_called {
			t.Errorf("got docker service called %d time, want %d", got_run_called, want_run_called)
		}

		var want_env_vars_map map[string]any
		err := json.Unmarshal(mock_app_cfg.VariablesJson, &want_env_vars_map)
		if err != nil {
			t.Fatal("got unexpected error unmarshaling test env vars", err)
		}

		got_run_called_with := stub_docker.run_call_args[0]
		want_run_called_with := docker.DockerRunDto{
			ImageName: mock_dp.ContainerImgName,
			ContainerPort: int(mock_instance.ContainerPort),
			ContainerName: mock_instance.ContainerName,
			EnvVars: want_env_vars_map,
		}
		if !reflect.DeepEqual(got_run_called_with, want_run_called_with) {
			t.Errorf("got docker run called with '%v', want '%v'", got_run_called_with, want_run_called_with)
		}

		// * Should populate deployment instance with values from docker service
		got_start_res := start_res
		want_start_res := mock_instance
		want_start_res.HostPort = int32(stub_docker.run_return.HostPort)		
		if !reflect.DeepEqual(got_start_res, want_start_res) {
			t.Errorf("got start result %v, want %v", got_start_res, want_start_res)
		}
	})

	// t.Run("should send message to the deployment channel on succeed", func (t * testing.T) {
	// 	defer func () {
	// 		stub_docker.Clear()
	// 	}()	

	// 	stub_docker.pull_error = nil
	// 	stub_docker.run_return = &docker.DockerRunResult{
	// 		HostPort: 9999,
	// 	}

	// 	mock_instance := database.DeploymentInstance{
	// 		ContainerName: "testapp-v1-abcd",
	// 		ContainerPort: mock_app_cfg.Port,
	// 	}

	// 	_, start_err := masbro_service.Start(deploy_request, mock_instance)
	// 	if start_err != nil {
	// 		t.Fatalf("got unexpected error '%v', want nil", start_err)
	// 	}
	// })
}