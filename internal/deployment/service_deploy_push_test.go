package deployment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/image"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/database"
	config "github.com/salmanrf/capybara-cloud/pkg/utils"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
)

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