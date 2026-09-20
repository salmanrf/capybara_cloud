package deployment

import (
	"log/slog"
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/application"
	config "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
)

func TestDeployStart(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	cfg := config.GetConfig()
	cfg.DOCKER_REGISTRY = "docker.io"
	cfg.DOCKER_NAMESPACE = "salmanrf"
	cfg.BASE_BUILD_PATH = "/tmp/masmasbro/builds"
	cfg.BASE_ARTIFACT_PATH = "/tmp/masmasbro/artifacts"
	config.SetConfig(cfg)

	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	docker := &StubDocker{}
	masbro_service := &StubMasbroService{}
	deployment_service := NewService(
		ctx,
		slog.Default(),
		docker,
		app_service,
		masbro_service,
		port_service,
		deployment_repository,
		make(chan DeployRequest, 1),
	)

	t.Run("should set error in DeployStepResult if createInstance fails", func (t *testing.T) {
		defer func (){
			masbro_service.Clear()
			deployment_repository.Clear()
		}() 
		
		want_error := errors.New("unable to create instance")
		deployment_repository.create_instance_error = want_error
		
		deploy_req := DeployRequest{
			ApplicationDto: database.Application{},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: database.ApplicationDeployment{},
		}

		res, _ := deployment_service.Start(deploy_req)
		got_error := res.DeploymentError
		if got_error == nil {
			t.Fatalf("got error '%v', want '%v'", got_error, want_error)
		}
		if !errors.Is(got_error, want_error) {
			t.Errorf("got error '%v', want wrapping '%v'", got_error, want_error)
		}

		got_create_ins_called := deployment_repository.create_instance_n_calls
		want_create_ins_called := 1
		if got_create_ins_called != want_create_ins_called {
			t.Errorf("got create instance called %d time, want %d", got_create_ins_called, want_create_ins_called)
		}
	})

	t.Run("should call createInstance with the correct values", func (t *testing.T) {
		defer func (){
			masbro_service.Clear()
			deployment_repository.Clear()
		}() 
		
		dep_id := pgtype.UUID{}
		dep_id.Scan(uuid.New().String())
		app_id := pgtype.UUID{}
		app_id.Scan(uuid.New().String())
		ins_id := pgtype.UUID{}
		ins_id.Scan(uuid.New().String())
		mock_dp := database.ApplicationDeployment{
			AppDpID: dep_id,
			ContainerRegistry: pgtype.Text{String: "docker.io", Valid: true},
			ContainerRepository: pgtype.Text{String: "masbro-cloud", Valid: true},
			ContainerTag: pgtype.Text{String: "001-2026-08-09-10-11-12", Valid: true},
		}
		mock_app := database.Application{
			AppID: app_id,
			Name: "masbro-cloud",
			Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
		}
		mock_cfg := database.ApplicationConfig{Port: pgtype.Int4{Int32: 8888, Valid: true}}
		deploy_req := DeployRequest{
			ApplicationDto: mock_app,
			ApplicationConfig: mock_cfg,
			DeploymentDto: mock_dp,
		}

		deployment_repository.create_instance_return = &database.DeploymentInstance{}
		masbro_service.start_error = errors.New("noop")

		deployment_service.Start(deploy_req)

		got_create_ins_called := deployment_repository.create_instance_n_calls
		want_create_ins_called := 1
		if got_create_ins_called != want_create_ins_called {
			t.Errorf("got create instance called %d time, want %d", got_create_ins_called, want_create_ins_called)
		}

		want_create_ins_params := database.CreateDeploymentInstanceParams{
			AppID: mock_app.AppID,
			DeploymentID: mock_dp.AppDpID,
			ContainerPort: mock_cfg.Port.Int32,
		}
		got_create_ins_params := deployment_repository.create_instance_call_args[0]
		if got_create_ins_params.AppID.String() != want_create_ins_params.AppID.String() {
			t.Errorf("got create instance called with app id '%v', want '%v'", got_create_ins_params.AppID, want_create_ins_params.AppID)
		}
		if got_create_ins_params.DeploymentID.String() != want_create_ins_params.DeploymentID.String() {
			t.Errorf("got create instance called with dep id '%v', want '%v'", got_create_ins_params.DeploymentID, want_create_ins_params.DeploymentID)
		}
		want_crate_ins_ct_pattern := regexp.MustCompile(fmt.Sprintf("%s-%s-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", mock_dp.ContainerRepository.String, mock_dp.ContainerTag.String, ))
		if !want_crate_ins_ct_pattern.Match([]byte(got_create_ins_params.ContainerName)) {
			t.Errorf("got create instance called with container name '%v', want matching '%s'", got_create_ins_params.ContainerName, want_crate_ins_ct_pattern.String())
		}
		if got_create_ins_params.ContainerPort != want_create_ins_params.ContainerPort {
			t.Errorf("got create instance called with container port '%v', want '%v'", got_create_ins_params.ContainerPort, want_create_ins_params.ContainerPort)
		}
		if got_create_ins_params.Status != want_create_ins_params.Status {
			t.Errorf("got create instance called with status '%v', want '%v'", got_create_ins_params.Status, want_create_ins_params.Status)
		}
	})

	t.Run("should set error in DeployStepResult if masbro_service.Start fails", func (t *testing.T) {
		defer func (){
			masbro_service.Clear()
			deployment_repository.Clear()
		}()

		want_error := errors.New("unable to start instance")
		masbro_service.start_error = want_error
		deployment_repository.create_instance_error = nil
		deployment_repository.create_instance_return = &database.DeploymentInstance{}

		deploy_req := DeployRequest{
			ApplicationDto: database.Application{},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: database.ApplicationDeployment{},
		}

		res, _ := deployment_service.Start(deploy_req)
		got_error := res.DeploymentError
		if got_error == nil {
			t.Fatalf("got error '%v', want '%v'", got_error, want_error)
		}
		if !errors.Is(got_error, want_error) {
			t.Errorf("got error '%v', want wrapping '%v'", got_error, want_error)
		}

		got_create_ins_called := deployment_repository.create_instance_n_calls
		want_create_ins_called := 1
		if got_create_ins_called != want_create_ins_called {
			t.Errorf("got create instance called %d time, want %d", got_create_ins_called, want_create_ins_called)
		}

		got_masbro_start_called := masbro_service.start_n_calls
		want_masbro_start_called := 1
		if got_masbro_start_called != want_masbro_start_called {
			t.Errorf("got masbro service.Start called %d time, want %d", got_masbro_start_called, want_masbro_start_called)
		}
	})

	t.Run("should set error in DeployStepResult if updateInstance fails", func (t *testing.T) {
		defer func (){
			masbro_service.Clear()
			deployment_repository.Clear()
		}()

		want_error := errors.New("unable to update instance")
		masbro_service.start_error = nil
		deployment_repository.create_instance_error = nil
		deployment_repository.create_instance_return = &database.DeploymentInstance{}
		deployment_repository.update_instance_error = want_error

		deploy_req := DeployRequest{
			ApplicationDto: database.Application{},
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: database.ApplicationDeployment{},
		}

		res, _ := deployment_service.Start(deploy_req)
		got_error := res.DeploymentError
		if got_error == nil {
			t.Fatalf("got error '%v', want '%v'", got_error, want_error)
		}
		if !errors.Is(got_error, want_error) {
			t.Errorf("got error '%v', want wrapping '%v'", got_error, want_error)
		}

		got_create_ins_called := deployment_repository.create_instance_n_calls
		want_create_ins_called := 1
		if got_create_ins_called != want_create_ins_called {
			t.Errorf("got create instance called %d time, want %d", got_create_ins_called, want_create_ins_called)
		}

		got_masbro_start_called := masbro_service.start_n_calls
		want_masbro_start_called := 1
		if got_masbro_start_called != want_masbro_start_called {
			t.Errorf("got masbro service.Start called %d time, want %d", got_masbro_start_called, want_masbro_start_called)
		}

		got_update_instance_called := deployment_repository.update_instance_n_calls
		want_update_instance_called := 1
		if got_update_instance_called != want_update_instance_called {
			t.Errorf("got masbro service.Start called %d time, want %d", got_update_instance_called, want_update_instance_called)
		}
	})

	t.Run("should call updateInstance with new values when no error", func (t *testing.T) {
		defer func (){
			masbro_service.Clear()
			deployment_repository.Clear()
		}()

		dep_id := pgtype.UUID{}
		dep_id.Scan(uuid.New().String())
		app_id := pgtype.UUID{}
		app_id.Scan(uuid.New().String())
		ins_id := pgtype.UUID{}
		ins_id.Scan(uuid.New().String())
		mock_dp := database.ApplicationDeployment{
			AppDpID: dep_id,
		}
		mock_app := database.Application{
			AppID: app_id,
		}
		deploy_req := DeployRequest{
			ApplicationDto: mock_app,
			ApplicationConfig: database.ApplicationConfig{},
			DeploymentDto: mock_dp,
		}

		want_instance := &database.DeploymentInstance{
			DeploymentID: mock_dp.AppDpID,
			AppID: mock_app.AppID,
			InstanceID: ins_id,
			Status: shared_deployment.DEPLOY_INSTANCE_STATUS_STOPPED,
		}
		masbro_service.start_error = nil
		masbro_service.start_return = *want_instance
		deployment_repository.create_instance_error = nil
		deployment_repository.create_instance_return = want_instance
		deployment_repository.update_instance_error = nil
		deployment_repository.update_instance_return = want_instance

		deployment_service.Start(deploy_req)

		got_create_ins_called := deployment_repository.create_instance_n_calls
		want_create_ins_called := 1
		if got_create_ins_called != want_create_ins_called {
			t.Errorf("got create instance called %d time, want %d", got_create_ins_called, want_create_ins_called)
		}

		got_masbro_start_called := masbro_service.start_n_calls
		want_masbro_start_called := 1
		if got_masbro_start_called != want_masbro_start_called {
			t.Errorf("got masbro service.Start called %d time, want %d", got_masbro_start_called, want_masbro_start_called)
		}

		got_update_instance_called := deployment_repository.update_instance_n_calls
		want_update_instance_called := 1
		if got_update_instance_called != want_update_instance_called {
			t.Errorf("got masbro service.Start called %d time, want %d", got_update_instance_called, want_update_instance_called)
		}

		got_start_return := masbro_service.start_return

		got_params := deployment_repository.update_instance_call_args[0]
		want_params := database.UpdateDeploymentInstanceParams{
			AppID: mock_app.AppID,
			DeploymentID: mock_dp.AppDpID,
			InstanceID: want_instance.InstanceID,
			Status: got_start_return.Status,
		}
		if got_params.InstanceID.String() != want_params.InstanceID.String() {
			t.Errorf("got instance id '%v', want '%v'", got_params.InstanceID, want_params.InstanceID)
		}
		if got_params.AppID.String() != want_params.AppID.String() {
			t.Errorf("got instance app id '%v', want '%v'", got_params.AppID, want_params.AppID)
		}
		if got_params.DeploymentID.String() != want_params.DeploymentID.String() {
			t.Errorf("got instance deployment id '%v', want '%v'", got_params.DeploymentID, want_params.DeploymentID)
		}
		if got_params.Status != want_params.Status {
			t.Errorf("got instance status id '%v', want '%v'", got_params.Status, want_params.Status)
		}
	})
}