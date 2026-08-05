package masbro_worker

import (
	"errors"
	"testing"

	"github.com/salmanrf/capybara-cloud/internal/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
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
		ArtifactsPath: "",
		BuildPath: "",
		VersionNumber: 1,
	}
	mock_app_cfg := database.ApplicationConfig{
		Port: 3000,
	}

	deploy_request := shared_deployment.DeployRequest{
		ApplicationDto: mock_app,
		DeploymentDto: mock_dp,
		ApplicationConfig: mock_app_cfg,
	}
	
	t.Run("should return error when docker pull fails", func (t *testing.T) {
		defer func () {
			stub_docker.Clear()
		}()	
		
		stub_docker.pull_error = errors.New("unable to pull image")

		_, err := masbro_service.Start(deploy_request)

		got_err := err
		want_err := stub_docker.pull_error
		if got_err == nil || got_err.Error() != want_err.Error() {
			t.Fatalf("got error %v, want %v", got_err, want_err)
		}
	})

	// t.Run("should start the container from pulled image", func (t * testing.T) {
	// 	defer func () {
	// 		stub_docker.Clear()
	// 	}()	
		
	// 	stub_docker.pull_error = nil
	// })
}