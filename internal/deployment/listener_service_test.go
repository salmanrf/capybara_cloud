package deployment

import (
	"testing"
	"time"

	"github.com/salmanrf/capybara-cloud/internal/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
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

		deployment_service.create_instance_return = &database.DeploymentInstance{}

		go listener_service.Listen()
		in_chan <- req

		timer := time.NewTimer(time.Second)

		select {
		case <- timer.C:
			t.Error("got unexpected timeout")
		case res := <- out_chan:
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

			want_update_instance_called := 1
			got_update_instance_called := deployment_service.update_instance_n_calls
			if got_update_instance_called != want_update_instance_called {
				t.Errorf("got deployment service update instancel called %d time, want %d", got_update_instance_called, want_update_instance_called)
			}

			want_new_deployment_status := shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED
			got_new_deployment_status := res.DeploymentDto.Status
			if got_new_deployment_status != int32(want_new_deployment_status) {
				t.Errorf("got new deployment status %d (DEPLOY_STATUS_BUILD_INSTANCE_STARTED), want %d", got_new_deployment_status, want_new_deployment_status)
			}
		}
	})
}