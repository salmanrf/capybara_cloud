package deployment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanrf/capybara-cloud/internal/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
) 

func createDeploymentServiceStub() *StubService {
	deployment_service := StubService{}
	deployment_service.extract_fn = func (dto shared_deployment.DeployRequest) (DeployStepResult, error) {
		res := DeployStepResult{
			ApplicationDto: dto.ApplicationDto,
			DeploymentDto: dto.DeploymentDto,
			ApplicationConfig: dto.ApplicationConfig,
		}
		return res, nil
	}
	deployment_service.build_fn = func (dto shared_deployment.DeployRequest) (DeployStepResult, error) {
		res := DeployStepResult{
			ApplicationDto: dto.ApplicationDto,
			DeploymentDto: dto.DeploymentDto,
			ApplicationConfig: dto.ApplicationConfig,
		}
		return res, nil
	}
	deployment_service.push_fn = func (dto shared_deployment.DeployRequest) (DeployStepResult, error) {
		res := DeployStepResult{
			ApplicationDto: dto.ApplicationDto,
			DeploymentDto: dto.DeploymentDto,
			ApplicationConfig: dto.ApplicationConfig,
		}
		return res, nil
	}
	deployment_service.create_instance_return = &database.DeploymentInstance{} 
	deployment_service.update_instance_return = &database.DeploymentInstance{}

	return &deployment_service
}

func TestDeployListener(t *testing.T) {
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
	deploy_step_res := DeployStepResult{
		ApplicationDto: mock_app,
		DeploymentDto: mock_dp,
		ApplicationConfig: mock_app_cfg,
	}

	t.Run("should stop processing input channel when context is cancelled", func (t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		in_chan := make(chan DeployRequest, 5)
		out_chan := make(chan DeployStepResult, 5)

		deployment_service := createDeploymentServiceStub()
		masbro_service := StubMasbroService{}
		listener_service := NewListener(ctx, in_chan, out_chan, deployment_service, &masbro_service)

		go listener_service.Listen()
		cancel()

		timer := time.NewTimer(500 * time.Millisecond)
		<- timer.C

		req := deploy_request
		in_chan <- req
		in_chan <- req
		in_chan <- req

		timer = time.NewTimer(500 * time.Millisecond)
		<- timer.C

		got_extract_called := deployment_service.extract_n_calls
		want_extract_called := 0
		if got_extract_called != want_extract_called {
			t.Fatalf("got extract called %d time, want %d", got_extract_called, want_extract_called)
		}
	})

	t.Run("should stop processing output channel when context is cancelled", func (t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		in_chan := make(chan DeployRequest, 5)
		out_chan := make(chan DeployStepResult, 5)

		deployment_service := createDeploymentServiceStub()
		masbro_service := StubMasbroService{}
		listener_service := NewListener(ctx, in_chan, out_chan, deployment_service, &masbro_service)

		go listener_service.Listen()
		cancel()

		timer := time.NewTimer(500 * time.Millisecond)
		<- timer.C

		req := deploy_step_res
		out_chan <- req
		out_chan <- req
		out_chan <- req

		timer = time.NewTimer(500 * time.Millisecond)
		<- timer.C

		got_update_called := deployment_service.update_n_calls
		want_update_called := 0
		if got_update_called != want_update_called {
			t.Fatalf("got update called %d time, want %d", got_update_called, want_update_called)
		}
	})

	t.Run("should call extract when status is DEPLOY_STATUS_INITIATED", func (t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in_chan := make(chan DeployRequest, 10)
		out_chan := make(chan DeployStepResult, 10)

		deployment_service := createDeploymentServiceStub()
		masbro_service := StubMasbroService{}
		listener_service := NewListener(ctx, in_chan, out_chan, deployment_service, &masbro_service)

		deployment_service.extract_fn = func(dto shared_deployment.DeployRequest) (res shared_deployment.DeployStepResult, err error) {
			// ? We care only about Extract, so listener can be cancelled after 
			// ? it had been called
			cancel()

			return res, err
		}

		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_INITIATED

		go listener_service.Listen()
		in_chan <- req 

		timer := time.NewTimer(500 * time.Millisecond)
		<- timer.C

		want_extract_called := 1
		got_extract_called := deployment_service.extract_n_calls

		if got_extract_called != want_extract_called {
			t.Errorf("got deployment_service.Extract called %d time, want 1", got_extract_called)
		}
	})
	
	t.Run("should call build when status is DEPLOY_STATUS_BUILD_EXTRACTED", func (t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in_chan := make(chan DeployRequest, 10)
		out_chan := make(chan DeployStepResult, 10)

		deployment_service := createDeploymentServiceStub()
		masbro_service := StubMasbroService{}
		listener_service := NewListener(ctx, in_chan, out_chan, deployment_service, &masbro_service)

		deployment_service.build_fn = func(dto shared_deployment.DeployRequest) (res shared_deployment.DeployStepResult, err error) {
			cancel()
			return res, err
		}

		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED
		
		go listener_service.Listen()
		in_chan <- req 

		timer := time.NewTimer(500 * time.Millisecond)
		<- timer.C

		want_build_called := 1
		got_build_called := deployment_service.build_n_calls

		if got_build_called != want_build_called {
			t.Errorf("got deployment_service.Build called %d time, want 1", got_build_called)
		}
	})

	t.Run("should call push when status is DEPLOY_STATUS_BUILD_IMAGE_BUILT", func (t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in_chan := make(chan DeployRequest, 10)
		out_chan := make(chan DeployStepResult, 10)

		deployment_service := createDeploymentServiceStub()
		masbro_service := StubMasbroService{}
		listener_service := NewListener(ctx, in_chan, out_chan, deployment_service, &masbro_service)

		deployment_service.build_fn = func(dto shared_deployment.DeployRequest) (res shared_deployment.DeployStepResult, err error) {
			cancel()
			return res, err
		}

		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		
		go listener_service.Listen()
		in_chan <- req 

		timer := time.NewTimer(500 * time.Millisecond)
		<- timer.C

		want_push_called := 1
		got_push_called := deployment_service.push_n_calls

		if got_push_called != want_push_called {
			t.Errorf("got deployment_service.Push called %d time, want 1", got_push_called)
		}
	})

	t.Run("should call start when status is DEPLOY_STATUS_BUILD_IMAGE_PUSHED", func (t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		in_chan := make(chan DeployRequest, 10)
		out_chan := make(chan DeployStepResult, 10)

		deployment_service := createDeploymentServiceStub()
		masbro_service := StubMasbroService{}
		listener_service := NewListener(ctx, in_chan, out_chan, deployment_service, &masbro_service)

		deployment_service.build_fn = func(dto shared_deployment.DeployRequest) (res shared_deployment.DeployStepResult, err error) {
			cancel()
			return res, err
		}

		req := deploy_request
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		
		go listener_service.Listen()
		in_chan <- req 

		timer := time.NewTimer(500 * time.Millisecond)
		<- timer.C

		want_start_called := 1
		got_start_called := deployment_service.start_n_calls
		if got_start_called != want_start_called {
			t.Errorf("got masbro_service.Start called %d time, want 1", got_start_called)
		}
	})
}

func TestDeployResultListener(t *testing.T) {
	mock_app := database.Application{}
	mock_dp := database.ApplicationDeployment{
		AppID: mock_app.AppID,
		Status: shared_deployment.DEPLOY_STATUS_INITIATED,
	}

	result := DeployStepResult{
		DeploymentDto: mock_dp,
		DeploymentError: nil,
	}

	t.Run("should update status when receiving errors", func (t *testing.T) {
		in_chan := make(chan DeployRequest)
		out_chan := make(chan DeployStepResult)

		masbro_service := StubMasbroService{}
		deployment_service := createDeploymentServiceStub()
		listener_service := NewListener(context.Background(), in_chan, out_chan, deployment_service, &masbro_service)

		defer func () {
			close(out_chan)
			close(in_chan)
		}()
		
		tests := []struct{
			status int
			want_status int
		}{
			{ 
				shared_deployment.DEPLOY_STATUS_INITIATED,
				-shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED,
			},
			{ 
				shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED,
				-shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT,
			},
			{ 
				shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT,
				-shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED,
			},
			{ 
				shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED,
				-shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED,
			},
		}

		for _, tt := range tests {
			t.Run("should update to the next status but negated", func (t *testing.T) {
				defer func () {
					deployment_service.Clear()
				}()

				req := result
				req.DeploymentDto.Status = int32(tt.status)

				want_error := errors.New("error")
				req.DeploymentError = want_error

				deployment_service.update_return = &req.DeploymentDto
				
				go listener_service.Listen()
				out_chan <- req

				timer := time.NewTimer(1 * time.Second)
				<- timer.C

				got_update_called := deployment_service.update_n_calls
				want_update_called := 1
				if got_update_called != want_update_called {
					t.Errorf("got deployment update called %d time, want %d", got_update_called, want_update_called)
				}

				got_new_status := deployment_service.update_call_args[0].Status
				want_new_status := tt.want_status
				if int(got_new_status) != int(want_new_status) {
					t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
				}
			})
		}
	})

	t.Run("should continue processing from DEPLOY_STATUS_INITIATED to DEPLOY_STATUS_BUILD_INSTANCE_STARTED", func (t *testing.T) {
		in_chan := make(chan DeployRequest)
		out_chan := make(chan DeployStepResult)

		masbro_service := StubMasbroService{}
		deployment_service := createDeploymentServiceStub()
		listener_service := NewListener(context.Background(), in_chan, out_chan, deployment_service, &masbro_service)

		defer func () {
			deployment_service.Clear()
			close(in_chan)
			close(out_chan)
		}()

		deployment_service.update_fn = func (dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
			updated := dep
			return &updated, nil
		}
		deployment_service.update_error = nil

		req := result
		// * Start from Extract output (Extract itself is not called)
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_INITIATED
		req.DeploymentError = nil

		go listener_service.Listen()
		out_chan <- req

		timer := time.NewTimer(time.Second)
		<- timer.C

		got_update_called := deployment_service.update_n_calls
		want_update_called := int(shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED - req.DeploymentDto.Status)
		if got_update_called != want_update_called {
			t.Fatalf("got deployment update called %d times, want %d", got_update_called, want_update_called)
		}

		got_build_called := deployment_service.build_n_calls
		want_build_called := 1
		if got_build_called != want_build_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_build_called, want_build_called)
		}
		got_push_called := deployment_service.push_n_calls
		want_push_called := 1
		if got_push_called != want_push_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_push_called, want_push_called)
		}
		got_create_instance_called := deployment_service.create_instance_n_calls
		want_create_instance_called := 1
		if got_create_instance_called != want_create_instance_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_create_instance_called, want_create_instance_called)
		}
		got_update_instance_called := deployment_service.update_instance_n_calls
		want_update_instance_called := 1
		if got_update_instance_called != want_update_instance_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_update_instance_called, want_update_instance_called)
		}
	})

	t.Run("should continue processing from DEPLOY_STATUS_BUILD_EXTRACTED to DEPLOY_STATUS_BUILD_INSTANCE_STARTED", func (t *testing.T) {
		in_chan := make(chan DeployRequest)
		out_chan := make(chan DeployStepResult)

		masbro_service := StubMasbroService{}
		deployment_service := createDeploymentServiceStub()
		listener_service := NewListener(context.Background(), in_chan, out_chan, deployment_service, &masbro_service)

		defer func () {
			deployment_service.Clear()
			close(in_chan)
			close(out_chan)
		}()

		deployment_service.update_fn = func (dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
			updated := dep
			return &updated, nil
		}
		deployment_service.update_error = nil

		req := result
		// * Start from Build output (Build itself is not called)
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED
		req.DeploymentError = nil

		go listener_service.Listen()
		out_chan <- req

		timer := time.NewTimer(time.Second)
		<- timer.C

		got_update_called := deployment_service.update_n_calls
		want_update_called := int(shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED - req.DeploymentDto.Status)
		if got_update_called != want_update_called {
			t.Fatalf("got deployment update called %d times, want %d", got_update_called, want_update_called)
		}

		got_push_called := deployment_service.push_n_calls
		want_push_called := 1
		if got_push_called != want_push_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_push_called, want_push_called)
		}
		got_create_instance_called := deployment_service.create_instance_n_calls
		want_create_instance_called := 1
		if got_create_instance_called != want_create_instance_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_create_instance_called, want_create_instance_called)
		}
		got_update_instance_called := deployment_service.update_instance_n_calls
		want_update_instance_called := 1
		if got_update_instance_called != want_update_instance_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_update_instance_called, want_update_instance_called)
		}
	})

	t.Run("should continue processing from DEPLOY_STATUS_BUILD_IMAGE_BUILT to DEPLOY_STATUS_BUILD_INSTANCE_STARTED", func (t *testing.T) {
		in_chan := make(chan DeployRequest)
		out_chan := make(chan DeployStepResult)

		masbro_service := StubMasbroService{}
		deployment_service := createDeploymentServiceStub()
		listener_service := NewListener(context.Background(), in_chan, out_chan, deployment_service, &masbro_service)

		defer func () {
			deployment_service.Clear()
			close(in_chan)
			close(out_chan)
		}()

		deployment_service.update_fn = func (dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
			updated := dep
			return &updated, nil
		}
		deployment_service.update_error = nil

		req := result
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		req.DeploymentError = nil

		go listener_service.Listen()
		out_chan <- req

		timer := time.NewTimer(time.Second)
		<- timer.C

		got_update_called := deployment_service.update_n_calls
		want_update_called := int(shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED - req.DeploymentDto.Status)
		if got_update_called != want_update_called {
			t.Fatalf("got deployment update called %d times, want %d", got_update_called, want_update_called)
		}

		got_create_instance_called := deployment_service.create_instance_n_calls
		want_create_instance_called := 1
		if got_create_instance_called != want_create_instance_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_create_instance_called, want_create_instance_called)
		}
		got_update_instance_called := deployment_service.update_instance_n_calls
		want_update_instance_called := 1
		if got_update_instance_called != want_update_instance_called {
			t.Fatalf("got deployment Extract called %d times, want %d", got_update_instance_called, want_update_instance_called)
		}
	})

	t.Run("should continue processing from DEPLOY_STATUS_BUILD_IMAGE_PUSHED to DEPLOY_STATUS_BUILD_INSTANCE_STARTED", func (t *testing.T) {
		in_chan := make(chan DeployRequest)
		out_chan := make(chan DeployStepResult)

		masbro_service := StubMasbroService{}
		deployment_service := createDeploymentServiceStub()
		listener_service := NewListener(context.Background(), in_chan, out_chan, deployment_service, &masbro_service)

		defer func () {
			deployment_service.Clear()
			close(in_chan)
			close(out_chan)
		}()

		deployment_service.update_fn = func (dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
			updated := dep
			return &updated, nil
		}
		deployment_service.update_error = nil

		req := result
		req.DeploymentDto.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED

		req.DeploymentError = nil

		go listener_service.Listen()
		out_chan <- req

		timer := time.NewTimer(time.Second)
		<- timer.C

		got_update_called := deployment_service.update_n_calls
		want_update_called := int(shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED - req.DeploymentDto.Status)
		if got_update_called != want_update_called {
			t.Fatalf("got deployment update called %d times, want %d", got_update_called, want_update_called)
		}
	})
}