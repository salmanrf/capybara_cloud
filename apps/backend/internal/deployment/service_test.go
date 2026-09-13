package deployment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/application"
	config "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

var mock_user_id = "df3e9f69-bbe4-4cbb-8479-dba87dd17e2c"
var mock_app_id = "d8dab736-e7a5-49f6-97a0-64670faa2a5d"

var mock_file_content = mock_file{false, []byte{10, 10, 10, 10}}
var mock_bundle = multipart.File(mock_file_content)

func TestNewService(t *testing.T) {
	t.Run("should setup directories on service creation", func (t *testing.T) {
		ctx := context.Background()
		deployment_repository := &StubAppDeploymentRepository{}
	
		deploy_chan := make(chan DeployRequest, 10)

		cfg := config.GetConfig()
		cfg.BASE_BUILD_PATH = "/tmp/tests/capybara-builds"
		cfg.BASE_ARTIFACT_PATH = "/tmp/tests/capybara-artifacts"
		config.SetConfig(cfg)
	
		port_service := &shared_deployment.StubPortAllocatorService{}
		app_service := &application.StubApplicationService{}
		NewService(
			ctx,
			&StubDocker{},
			app_service,
			&StubMasbroService{},
			port_service,
			deployment_repository,
			deploy_chan,
		)

		_, err := os.ReadDir(cfg.BASE_BUILD_PATH)
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}
		_, err = os.ReadDir(cfg.BASE_ARTIFACT_PATH)
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}
	})

	t.Run("should panic if directories setup failed", func (t *testing.T) {
		ctx := context.Background()
		deployment_repository := &StubAppDeploymentRepository{}
	
		deploy_chan := make(chan DeployRequest, 10)

		cfg := config.GetConfig()
		cfg.BASE_BUILD_PATH = "/111"
		cfg.BASE_ARTIFACT_PATH = "/abcd"
		config.SetConfig(cfg)
	
		port_service := &shared_deployment.StubPortAllocatorService{}
		app_service := &application.StubApplicationService{}

		defer func () {
			if err := recover(); err == nil {
				t.Errorf("got error nil, want error")
			}
		}()

		NewService(
			ctx,
			&StubDocker{},
			app_service,
			&StubMasbroService{},
			port_service,
			deployment_repository,
			deploy_chan,
		)
	})
}

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	deploy_chan := make(chan DeployRequest, 10)

	cfg := config.GetConfig()	
	cfg.BASE_ARTIFACT_PATH = "/tmp/tests/artifacts"
	cfg.BASE_BUILD_PATH = "/tmp/tests/builds"
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
		deploy_chan,
	)

	t.Run("should return error if application is not found", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle);

		got_dp := dp
		var want_dp *database.ApplicationDeployment = nil
		want_error := errors.New("not_found")
		got_error := err

		if got_dp != want_dp {
			t.Errorf("got app deployment %v, want %v", got_dp, want_dp)
		}
		
		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})

	t.Run("should return error if the logged in user is not a project member", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()

		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		app_service.Find_one_return = mock_app

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle);

		got_dp := dp
		var want_dp *database.ApplicationDeployment = nil
		want_error := errors.New("permission_denied")
		got_error := err

		if got_dp != want_dp {
			t.Errorf("got app deployment %v, want %v", got_dp, want_dp)
		}
		
		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})

	t.Run("should return error if application config is missing", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()

		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "container_nodejs"
		mock_app.ProjectMember.ProjectID.Scan(uuid.New().String())
		app_service.Find_one_return = mock_app

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle);

		got_dp := dp
		var want_dp *database.ApplicationDeployment = nil
		want_error := errors.New("config_not_found")
		got_error := err

		if got_dp != want_dp {
			t.Errorf("got app deployment %v, want %v", got_dp, want_dp)
		}

		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}

		if n := len(deployment_repository.create_call_args); n != 0 {
			t.Errorf("got %d deployment create calls, want 0", n)
		}

		if n := len(deploy_chan); n != 0 {
			t.Errorf("got %d queued deploy requests, want 0", n)
		}
	})

	t.Run("should create and return the application deployment db item", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<- deploy_chan
			}
		}()
		
		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID = pgtype.UUID{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "container_nodejs"
		mock_app.ApplicationConfig = database.ApplicationConfig{}
		mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ProjectMember.ProjectID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

		mock_current_dp := &database.ApplicationDeployment{VersionNumber: 4}

		want_create_dp_params := database.CreateApplicationDeploymentParams{
			AppID: mock_app.AppID,
			StorageService: "localfs",
			VariablesSnapshotJson: mock_app.ApplicationConfig.VariablesJson,
			VersionNumber: mock_current_dp.VersionNumber + 1,
			Status: shared_deployment.DEPLOY_STATUS_INITIATED,
			ContainerRegistry: pgtype.Text{String: cfg.DOCKER_REGISTRY, Valid: true},
			ContainerNamespace: pgtype.Text{String: cfg.DOCKER_NAMESPACE, Valid: true},
			ContainerRepository: pgtype.Text{String: utils.Slugify(mock_app.Name), Valid: true},
		}

		app_service.Find_one_return = mock_app

		expected_deployment := &database.ApplicationDeployment{
			AppID: mock_app.AppID,
			ArtifactsPath: "/tmp/tests/artifacts/artifact-handsome-capybara-2026-01-02-03-04-05/abcd.tar.gz",
			BuildPath: "/tmp/tests/builds/build-handsome-capybara-2026-01-02-03-04-05",
			VariablesSnapshotJson: mock_app.ApplicationConfig.VariablesJson,
			StorageService: "localfs",
			VersionNumber: 5,
			Status: shared_deployment.DEPLOY_STATUS_INITIATED,
			ContainerRegistry: pgtype.Text{String: cfg.DOCKER_REGISTRY, Valid: true},
			ContainerNamespace: pgtype.Text{String: cfg.DOCKER_NAMESPACE, Valid: true},
			ContainerRepository: pgtype.Text{String: "handsome-capybara", Valid: true},
			ContainerTag: pgtype.Text{String: "005-2026-01-02-03-04-05", Valid: true},
		}
		expected_deployment.AppDpID.Scan(uuid.New().String())
		expected_deployment.CreatedAt.Scan(time.Now())
		expected_deployment.UpdatedAt.Scan(time.Now())

		deployment_repository.find_current_return = mock_current_dp
		deployment_repository.create_return = expected_deployment

		dp, err := deployment_service.Deploy(
			mock_user_id,
			mock_app_id,
			&multipart.FileHeader{
				Filename: "abcd.tar.gz",
			},
			mock_bundle,
		);

		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}

		if dp == nil {
			t.Fatal("got deployment nil, want pointer")
		}

		if diff := cmp.Diff(expected_deployment, dp); diff != "" {
			t.Errorf("deployment mismatch (-want +got):\n%s", diff)
		}

		if n := len(deployment_repository.create_call_args); n != 1 {
			t.Fatalf("got %d deployment create calls, want 1", n)
		}

		got_create_dp_params := deployment_repository.create_call_args[0]

		ts_pattern := "\\d{4}-\\d{2}-\\d{2}-\\d{2}-\\d{2}-\\d{2}"
		slug := utils.Slugify(mock_app.Name)

		artifact_path_pattern := regexp.MustCompile(fmt.Sprintf("^%s/artifact-%s-%s/abcd\\.tar\\.gz$", cfg.BASE_ARTIFACT_PATH, slug, ts_pattern))
		if !artifact_path_pattern.MatchString(got_create_dp_params.ArtifactsPath) {
			t.Errorf("got artifacts_path %q, want matching %s", got_create_dp_params.ArtifactsPath, artifact_path_pattern)
		}

		build_path_pattern := regexp.MustCompile(fmt.Sprintf("^%s/build-%s-%s$", cfg.BASE_BUILD_PATH, slug, ts_pattern))
		if !build_path_pattern.MatchString(got_create_dp_params.BuildPath) {
			t.Errorf("got build_path %q, want matching %s", got_create_dp_params.BuildPath, build_path_pattern)
		}

		container_tag_pattern := regexp.MustCompile(fmt.Sprintf("^%03d-%s$", want_create_dp_params.VersionNumber, ts_pattern))
		if !got_create_dp_params.ContainerTag.Valid || !container_tag_pattern.MatchString(got_create_dp_params.ContainerTag.String) {
			t.Errorf("got container_tag %q (valid=%v), want matching %s", got_create_dp_params.ContainerTag.String, got_create_dp_params.ContainerTag.Valid, container_tag_pattern)
		}

		got_create_dp_params.ArtifactsPath = ""
		got_create_dp_params.BuildPath = ""
		got_create_dp_params.ContainerTag = pgtype.Text{}

		if diff := cmp.Diff(want_create_dp_params, got_create_dp_params); diff != "" {
			t.Errorf("create deployment params mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should pass a message to deploy_chan on successful creation", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()

		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.ProjectID.Scan(uuid.New().String())
		mock_app.ProjectMember.ProjectID.Scan(uuid.New().String())
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "container_nodejs"
		mock_app.CreatedAt.Scan(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
		mock_app.UpdatedAt.Scan(time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC))
		mock_app.ProjectMember.Role = pgtype.Text{String: "owner", Valid: true}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.AppID = mock_app.AppID
		mock_app.ApplicationConfig.VariablesJson = []byte(`{"PORT":"8080"}`)
		mock_app.ApplicationConfig.Port = pgtype.Int4{Int32: 8080, Valid: true}
		mock_app.ApplicationConfig.RuntimeType = pgtype.Text{String: "nodejs", Valid: true}
		mock_app.ApplicationConfig.CreatedAt.Scan(time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC))
		mock_app.ApplicationConfig.UpdatedAt.Scan(time.Date(2026, 4, 5, 6, 7, 8, 0, time.UTC))

		want_app_dto := database.Application{
			AppID: mock_app.AppID,
			ProjectID: mock_app.ProjectID,
			Type: mock_app.Type,
			Name: mock_app.Name,
			CreatedAt: mock_app.CreatedAt,
			UpdatedAt: mock_app.UpdatedAt,
		}
		want_app_config_dto := database.ApplicationConfig{
			AppCfgID: mock_app.ApplicationConfig.AppCfgID,
			AppID: mock_app.ApplicationConfig.AppID,
			VariablesJson: mock_app.ApplicationConfig.VariablesJson,
			Port: mock_app.ApplicationConfig.Port,
			RuntimeType: mock_app.ApplicationConfig.RuntimeType,
			CreatedAt: mock_app.ApplicationConfig.CreatedAt,
			UpdatedAt: mock_app.ApplicationConfig.UpdatedAt,
		}

		app_service.Find_one_return = mock_app

		expected_deployment := &database.ApplicationDeployment{
			AppID: mock_app.AppID,
			ArtifactsPath: "/tmp/tests/artifacts/artifact-handsome-capybara-2026-01-02-03-04-05/abcd.tar.gz",
			BuildPath: "/tmp/tests/builds/build-handsome-capybara-2026-01-02-03-04-05",
			VariablesSnapshotJson: mock_app.ApplicationConfig.VariablesJson,
			StorageService: "localfs",
			VersionNumber: 1,
			Status: shared_deployment.DEPLOY_STATUS_INITIATED,
			ContainerRegistry: pgtype.Text{String: cfg.DOCKER_REGISTRY, Valid: true},
			ContainerNamespace: pgtype.Text{String: cfg.DOCKER_NAMESPACE, Valid: true},
			ContainerRepository: pgtype.Text{String: "handsome-capybara", Valid: true},
			ContainerTag: pgtype.Text{String: "001-2026-01-02-03-04-05", Valid: true},
		}
		expected_deployment.AppDpID.Scan(uuid.New().String())
		expected_deployment.CreatedAt.Scan(time.Now())
		expected_deployment.UpdatedAt.Scan(time.Now())

		deployment_repository.create_return = expected_deployment

		timer := time.NewTimer(3 * time.Second)
		dp, err := deployment_service.Deploy(
			mock_user_id,
			mock_app_id,
			&multipart.FileHeader{
				Filename: "abcd.tar.gz",
			},
			mock_bundle,
		);

		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}

		if diff := cmp.Diff(expected_deployment, dp); diff != "" {
			t.Errorf("deployment mismatch (-want +got):\n%s", diff)
		}

		want_message := DeployRequest{
			ApplicationDto: want_app_dto,
			ApplicationConfig: want_app_config_dto,
			DeploymentDto: *expected_deployment,
		}

		select {
		case <- timer.C:
			t.Fatal("Deploy timeout reached!")
		case got_message := <- deploy_chan:
			if diff := cmp.Diff(want_message, got_message); diff != "" {
				t.Errorf("deploy request mismatch (-want +got):\n%s", diff)
			}
		}

		if n := len(deploy_chan); n != 0 {
			t.Errorf("got %d extra queued deploy requests, want 0", n)
		}
	})

	t.Run("should return error if find current deployment fails", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()
		
		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID = pgtype.UUID{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "container_nodejs"
		mock_app.ApplicationConfig = database.ApplicationConfig{}
		mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ProjectMember.ProjectID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

		app_service.Find_one_return = mock_app

		expected_deployment := &database.ApplicationDeployment{
			AppID: mock_app.AppID,
		}
		expected_deployment.AppDpID.Scan(uuid.New())

		deployment_repository.create_return = expected_deployment
		deployment_repository.find_current_error = errors.New("Invalid something")

		_, got_err := deployment_service.Deploy(mock_user_id, mock_app_id, &multipart.FileHeader{}, mock_bundle);

		if got_err == nil {
			t.Errorf("got error nil, want error")
		}
	})

	t.Run("should defaults version number to 1 if find current deployment doesn't exist", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()
		
		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID = pgtype.UUID{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "container_nodejs"
		mock_app.ApplicationConfig = database.ApplicationConfig{}
		mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ProjectMember.ProjectID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

		app_service.Find_one_return = mock_app

		expected_deployment := &database.ApplicationDeployment{
			AppID: mock_app.AppID,
		}
		expected_deployment.AppDpID.Scan(uuid.New())

		deployment_repository.create_return = expected_deployment
		deployment_repository.find_current_return = nil
		deployment_repository.find_current_error = errors.New("not_found")

		deployment_service.Deploy(
			mock_user_id, 
			mock_app_id, 
			&multipart.FileHeader{
				Filename: "zsh.tar.gz",
			}, 
			mock_bundle,
		);

		expected_create_dp_arg := deployment_repository.create_call_args[0]
		
		got_version_number := expected_create_dp_arg.VersionNumber
		want_version_number := int32(1)

		if got_version_number != want_version_number {
			t.Errorf("got deployment version number %v, want %v", got_version_number, want_version_number)
		}
	})

	t.Run("should auto-increment version from the current deployment", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
			for len(deploy_chan) > 0 {
				<-deploy_chan
			}
		}()
		
		mock_current_dp := &database.ApplicationDeployment{}
		mock_current_dp.VersionNumber = 10
		
		mock_app := &database.FindOneApplicationCompleteRow{}
		mock_app.AppID = pgtype.UUID{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "container_nodejs"
		mock_app.ApplicationConfig = database.ApplicationConfig{}
		mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ProjectMember.ProjectID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

		app_service.Find_one_return = mock_app

		expected_deployment := &database.ApplicationDeployment{
			AppID: mock_app.AppID,
		}
		expected_deployment.AppDpID.Scan(uuid.New())

		deployment_repository.create_return = expected_deployment
		deployment_repository.find_current_return = mock_current_dp

		deployment_service.Deploy(
			mock_user_id, 
			mock_app_id, 
			&multipart.FileHeader{
				Filename: "asd.tar.gz",
			}, 
			mock_bundle,
		);

		expected_create_dp_arg := deployment_repository.create_call_args[0]
		
		got_version_number := expected_create_dp_arg.VersionNumber
		want_version_number := mock_current_dp.VersionNumber + 1

		if got_version_number != want_version_number {
			t.Errorf("got deployment version number %v, want %v", got_version_number, want_version_number)
		}
	})
}

func TestCreateInstance(t *testing.T) {
	ctx := context.Background()

	deploy_chan := make(chan DeployRequest, 10)
	
	deployment_repository := &StubAppDeploymentRepository{}
	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		&StubDocker{},
		app_service,
		&StubMasbroService{},
		port_service,
		deployment_repository,
		deploy_chan,
	)

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
		ContainerRepository: pgtype.Text{String: "testapp", Valid: true},
		ContainerTag: pgtype.Text{String: "001-2026-08-09-10-11-12", Valid: true},
	}
	mock_app_cfg := database.ApplicationConfig{
		Port: pgtype.Int4{Int32: 3000, Valid: true},
	}

	deploy_request := DeployRequest{
		ApplicationDto: mock_app,
		DeploymentDto: mock_dp,
		ApplicationConfig: mock_app_cfg,
	}
	
	t.Run("should construct deployment instance from DeployRequest", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
		}()

		expected_instance := &database.DeploymentInstance{
			AppID: mock_app.AppID,
			DeploymentID: mock_dp.AppDpID,
			ContainerPort: mock_app_cfg.Port.Int32,
		}
		expected_instance.InstanceID.Scan(uuid.New().String())

		deployment_repository.create_instance_return = expected_instance

		ins, err := deployment_service.createInstance(deploy_request)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_instance := *ins
		want_instance := *expected_instance

		if !reflect.DeepEqual(got_instance, want_instance) {
			t.Errorf("got instance %v, want %v", got_instance, want_instance)
		}
	})

	t.Run("should call repository CreateInstance with params from DeployRequest", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
		}()

		deployment_repository.create_instance_return = &database.DeploymentInstance{}

		_, err := deployment_service.createInstance(deploy_request)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_n_calls := deployment_repository.create_instance_n_calls
		want_n_calls := 1

		if got_n_calls != want_n_calls {
			t.Fatalf("got CreateInstance called %d times, want %d", got_n_calls, want_n_calls)
		}

		got_create_ins_arg := deployment_repository.create_instance_call_args[0]

		if got_create_ins_arg.AppID != mock_app.AppID {
			t.Errorf("got app id %v, want %v", got_create_ins_arg.AppID, mock_app.AppID)
		}
		if got_create_ins_arg.DeploymentID != mock_dp.AppDpID {
			t.Errorf("got deployment id %v, want %v", got_create_ins_arg.DeploymentID, mock_dp.AppDpID)
		}
		if got_create_ins_arg.ContainerPort != mock_app_cfg.Port.Int32 {
			t.Errorf("got container port %v, want %v", got_create_ins_arg.ContainerPort, mock_app_cfg.Port)
		}

		pattern := regexp.MustCompile(fmt.Sprintf("^%s-%s-[0-9a-f-]{36}$", mock_dp.ContainerRepository.String, mock_dp.ContainerTag.String))
		if !pattern.MatchString(got_create_ins_arg.ContainerName) {
			t.Errorf("got container name %s, want matching %s", got_create_ins_arg.ContainerName, pattern.String())
		}
	})

	t.Run("should return error from repository CreateInstance", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
		}()

		deployment_repository.create_instance_error = errors.New("create_instance_failed")

		ins, err := deployment_service.createInstance(deploy_request)
		if err == nil {
			t.Fatalf("got nil error, want %v", deployment_repository.create_instance_error)
		}
		if ins != nil {
			t.Errorf("got instance %v, want nil", ins)
		}
	})
}

func TestUpdateInstance(t *testing.T) {
	ctx := context.Background()

	deploy_chan := make(chan DeployRequest, 10)

	deployment_repository := &StubAppDeploymentRepository{}
	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		&StubDocker{},
		app_service,
		&StubMasbroService{},
		port_service,
		deployment_repository,
		deploy_chan,
	)

	mock_instance := &database.DeploymentInstance{
		ContainerName: "container-nodejs-testapp-2026-08-09-10-10-10",
		HostPort: 42690,
		ContainerPort: 3000,
	}
	mock_instance.InstanceID.Scan(uuid.New().String())
	mock_instance.AppID.Scan(mock_app_id)
	mock_instance.DeploymentID.Scan(uuid.New().String())

	t.Run("should call repository UpdateInstance with params from the instance", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
		}()

		deployment_repository.update_instance_return = &database.DeploymentInstance{}

		_, err := deployment_service.updateInstance(mock_instance)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_n_calls := deployment_repository.update_instance_n_calls
		want_n_calls := 1

		if got_n_calls != want_n_calls {
			t.Fatalf("got UpdateInstance called %d times, want %d", got_n_calls, want_n_calls)
		}

		got_update_ins_arg := deployment_repository.update_instance_call_args[0]
		want_update_ins_arg := database.UpdateDeploymentInstanceParams{
			InstanceID: mock_instance.InstanceID,
			AppID: mock_instance.AppID,
			DeploymentID: mock_instance.DeploymentID,
			ContainerName: mock_instance.ContainerName,
			HostPort: mock_instance.HostPort,
			ContainerPort: mock_instance.ContainerPort,
		}

		if !reflect.DeepEqual(got_update_ins_arg, want_update_ins_arg) {
			t.Errorf("got update instance params %v, want %v", got_update_ins_arg, want_update_ins_arg)
		}
	})

	t.Run("should return the updated instance from repository UpdateInstance", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
		}()

		expected_instance := &database.DeploymentInstance{
			InstanceID: mock_instance.InstanceID,
			AppID: mock_instance.AppID,
			DeploymentID: mock_instance.DeploymentID,
			ContainerName: mock_instance.ContainerName,
			HostPort: mock_instance.HostPort,
			ContainerPort: mock_instance.ContainerPort,
		}

		deployment_repository.update_instance_return = expected_instance

		ins, err := deployment_service.updateInstance(mock_instance)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_instance := *ins
		want_instance := *expected_instance

		if !reflect.DeepEqual(got_instance, want_instance) {
			t.Errorf("got instance %v, want %v", got_instance, want_instance)
		}
	})

	t.Run("should return error from repository UpdateInstance", func (t *testing.T) {
		defer func () {
			app_service.Clear()
			deployment_repository.Clear()
		}()

		deployment_repository.update_instance_error = errors.New("update_instance_failed")

		ins, err := deployment_service.updateInstance(mock_instance)
		if err == nil {
			t.Fatalf("got nil error, want %v", deployment_repository.update_instance_error)
		}
		if ins != nil {
			t.Errorf("got instance %v, want nil", ins)
		}
	})
}

func TestGetFullArtifactPath(t *testing.T) {
	config := config.Config{}
	config.BASE_ARTIFACT_PATH = "/tmp/test_get_full_dir_path"
	
	mock_app := database.FindOneApplicationCompleteRow{}
	mock_app.AppID = pgtype.UUID{}
	mock_app.AppID.Scan(mock_app_id)
	mock_app.Type = "container_nodejs"
	mock_app.ApplicationConfig = database.ApplicationConfig{}
	mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
	mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
	mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

	tests := []struct{
		appname string
		fileheaders *multipart.FileHeader
	}{
		{"Handsome Capybara", &multipart.FileHeader{Filename: "capybundle.tar.gz"}},
		{"123Tailung&Von Astrea456__greyR4t", &multipart.FileHeader{Filename: "bundle.tar.gz"}},
		{"Sophia Foundations FE", &multipart.FileHeader{Filename: "sophsoph.tar.gz"}},
		{"Selma Sharia Finance", &multipart.FileHeader{Filename: "selmafin.tar.gz"}},
	}

	want_basepath := config.BASE_ARTIFACT_PATH + "/artifact"

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should format app name '%s' into artifact path", tt.appname), func (t *testing.T) {
			mock_app.Name = tt.appname
			appnameslug := utils.Slugify(tt.appname)

			got_path := getArtifactDirPath(config, "localfs", mock_app)

			pattern, _ := regexp.Compile(fmt.Sprintf("%s-%s-\\d{4}-\\d{2}-\\d{2}-\\d{2}-\\d{2}-\\d{2}", want_basepath, appnameslug))
			if !pattern.Match([]byte(got_path)) {
				t.Errorf("got match false (%s), want starting with %s", got_path, want_basepath)
			}
		})
	}
}

func TestGetContainerTag(t *testing.T) {
	tests := []struct{
		name string
		version int
		created_at time.Time
	}{
		{
			"capybara masbro", 
			1, 
			time.Now(),
		},
		{
			"sophia shops", 
			10,
			time.Now().AddDate(0, -6, 0),
		},
		{
			"capybara web services", 
			5,
			time.Now().AddDate(-1, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should return the docker repo tag for '%s'", tt.name), func (t *testing.T) {
			created_at := pgtype.Timestamp{}
			created_at.Scan(time.Now())
			mock_dp := database.ApplicationDeployment{
				VersionNumber: int32(tt.version),
				CreatedAt: created_at,
			}
			tag := fmt.Sprintf(
				"%03s-%s", 
				fmt.Sprintf("%d", mock_dp.VersionNumber),
				utils.DockerSafeDateString(mock_dp.CreatedAt.Time), 
			)

			format := fmt.Sprintf("^%s$", tag)
			want_pattern, err := regexp.Compile(format)
			if err != nil {
				t.Fatal(err)
			}

			got_str := getContainerTag(mock_dp)
			if got_match := want_pattern.Match([]byte(got_str)); !got_match {
				t.Errorf("got container tag %s, want matching %s", got_str, want_pattern.String())
			}
		})
	}
	
}

func TestGetContainerName(t *testing.T) {
	tests := []struct{
		repository string
		tag string
	}{
		{"handsome-capybara", "001-2026-08-09-10-11-12"},
		{"123tailung-von-astrea456-greyr4t", "002-2026-08-09-10-11-12"},
		{"sophia-foundations-fe", "003-2026-08-09-10-11-12"},
		{"selma-sharia-finance", "010-2026-08-09-10-11-12"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should format '%s:%s' into container name", tt.repository, tt.tag), func (t *testing.T) {
			mock_dp := database.ApplicationDeployment{
				ContainerRepository: pgtype.Text{String: tt.repository, Valid: true},
				ContainerTag: pgtype.Text{String: tt.tag, Valid: true},
			}

			got_name := getContainerName(mock_dp)

			pattern := regexp.MustCompile(fmt.Sprintf("^%s-%s-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", tt.repository, tt.tag))
			if !pattern.MatchString(got_name) {
				t.Errorf("got container name %s, want matching %s", got_name, pattern.String())
			}
		})
	}
}

func TestSaveDeployArtifacts(t *testing.T) {
	config := config.Config{
		BASE_ARTIFACT_PATH: "/tmp/test_save_artifact",
	}
	
	mock_app := &database.FindOneApplicationCompleteRow{}
	mock_app.AppID = pgtype.UUID{}
	mock_app.AppID.Scan(mock_app_id)
	mock_app.Type = "container_nodejs"
	mock_app.ApplicationConfig = database.ApplicationConfig{}
	mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
	mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
	mock_app.ApplicationConfig.VariablesJson = []byte(`JWT_SECRET=abcd`)

	tests := []struct{
		appname string
		port int
		samplefilename string
	}{
		{"Masbro Capybara", 3000, "node-express.tar.gz"},
		{"Tailung---GreyR4ttts", 5050, "node-express.tar.gz"},
		{"Sophia Foundations Landing Page", 8080, "node-express.tar.gz"},
		{"Selman$$$Finance", 8888, "node-express.tar.gz"},
	}

	pwd, _ := os.Getwd()

	for _, tt := range tests {
		t.Run("should create the artifact directory and return the path given an app and bundle file", func (t *testing.T) {			
			mock_app := *mock_app
			mock_app.Name = tt.appname
					
			samplefile, _ := os.OpenFile(path.Join(pwd, "samples", tt.samplefilename), os.O_RDONLY, 0) 
			fstat, _ := samplefile.Stat()
			fheaders := &multipart.FileHeader{
				Filename: fstat.Name(),
				Size: fstat.Size(),
			}
			
			got_dir_path := getArtifactDirPath(config, "", mock_app)			
			want_full_path := got_dir_path + "/" + tt.samplefilename
			got_path, got_err := saveDeployArtifacts(config, mock_app, fheaders, samplefile)

			defer func () {
				samplefile.Close()
				err := os.RemoveAll(config.BASE_ARTIFACT_PATH)
				if err != nil {
					fmt.Println("Cleanup error:", err)
				}
			}()

			if got_err != nil {
				t.Errorf("got error %v, want nil", got_err)
			}

			if got_path != want_full_path {
				t.Errorf("got artifact path %s, want %s", got_path, got_dir_path)
			}

			_, err := os.Stat(got_dir_path)
			if err != nil {
				t.Errorf("got artifact dir stat error %v, want nil", err)
			}

			_, err = os.Stat(want_full_path)
			if err != nil {
				t.Errorf("got bundle artifact stat error %v, want nil", err)
			}
		})
	}
}

func TestGetBuildDirPath(t *testing.T) {
	cfg := config.GetConfig()
	cfg.BASE_BUILD_PATH = "/tmp/tests/build"
	
	tests := []struct{
		name string
	}{
		{"capybara-heyheyhey"},
		{"Sophia One Commerce "},
		{"Selma Mobile Banking 123"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should build path for build step, starting with %s/<app_name_slug>-<timestamp>", cfg.BASE_BUILD_PATH), func (t *testing.T) {
			app_name := tt.name
			slug := utils.Slugify(app_name)
			got_build_path := getBuildDirPath(
				cfg,
				"localfs",
				app_name,
			)
			
			pattern, _ := regexp.Compile(fmt.Sprintf("%s/build-%s-\\d{4}-\\d{2}-\\d{2}-\\d{2}-\\d{2}-\\d{2}", cfg.BASE_BUILD_PATH, slug))
			if !pattern.Match([]byte(got_build_path)) {
				t.Errorf("got match false (%s), want true", got_build_path)
			}
		})
	}
}

type mock_file struct {
	hasMore bool
	buf []byte
}

type MockFile interface {
	multipart.File
}

func (m mock_file) Read(buf []byte) (int, error) {
	if !m.hasMore {
		return 0, io.EOF
	}
	
	bp := &buf
	*bp = m.buf
	
	m.hasMore = true
	
	return len(m.buf), nil
}
func (m mock_file) ReadAt(buf []byte, _ int64) (int, error) {
	bp := &buf
	*bp = m.buf
	return len(m.buf), nil
}
func (m mock_file) Seek(int64, int) (n int64, e error) { return n, e }
func (m mock_file) Close() error { return nil }
func TestMain(m *testing.M) {
	config.CreateLogger()
	os.Exit(m.Run())
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	deploy_chan := make(chan DeployRequest, 10)

	deployment_repository := &StubAppDeploymentRepository{}
	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		&StubDocker{},
		app_service,
		&StubMasbroService{},
		port_service,
		deployment_repository,
		deploy_chan,
	)

	mock_dep := database.ApplicationDeployment{
		ArtifactsPath: "/tmp/artifact-testapp/bundle.tar.gz",
		VariablesSnapshotJson: []byte(`{"PORT":"3000"}`),
		StorageService: "localfs",
		VersionNumber: 3,
		Status: shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT,
		BuildPath: "/tmp/build-testapp",
		ContainerRegistry: pgtype.Text{String: "docker.io", Valid: true},
		ContainerNamespace: pgtype.Text{String: "testreg", Valid: true},
		ContainerRepository: pgtype.Text{String: "testapp", Valid: true},
		ContainerTag: pgtype.Text{String: "2026-08-09-003", Valid: true},
	}
	mock_dep.AppDpID.Scan(uuid.New().String())
	mock_dep.AppID.Scan(mock_app_id)

	t.Run("should call repository Update with params from the deployment", func (t *testing.T) {
		defer deployment_repository.Clear()

		deployment_repository.update_return = &database.ApplicationDeployment{}

		_, err := deployment_service.update(mock_dep)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		got_n_calls := deployment_repository.update_n_calls
		want_n_calls := 1
		if got_n_calls != want_n_calls {
			t.Fatalf("got Update called %d times, want %d", got_n_calls, want_n_calls)
		}

		got_update_arg := deployment_repository.update_call_args[0]
		want_update_arg := database.UpdateDeploymentParams{
			AppDpID: mock_dep.AppDpID,
			AppID: mock_dep.AppID,
			ArtifactsPath: mock_dep.ArtifactsPath,
			VariablesSnapshotJson: mock_dep.VariablesSnapshotJson,
			StorageService: mock_dep.StorageService,
			VersionNumber: mock_dep.VersionNumber,
			Status: mock_dep.Status,
			BuildPath: mock_dep.BuildPath,
			ContainerRegistry: mock_dep.ContainerRegistry,
			ContainerNamespace: mock_dep.ContainerNamespace,
			ContainerRepository: mock_dep.ContainerRepository,
			ContainerTag: mock_dep.ContainerTag,
		}

		if !reflect.DeepEqual(got_update_arg, want_update_arg) {
			t.Errorf("got update params %v, want %v", got_update_arg, want_update_arg)
		}
	})

	t.Run("should return the updated deployment from repository Update", func (t *testing.T) {
		defer deployment_repository.Clear()

		expected_dep := mock_dep
		deployment_repository.update_return = &expected_dep

		dep, err := deployment_service.update(mock_dep)
		if err != nil {
			t.Fatalf("got unexpected error %v, want nil", err)
		}

		if !reflect.DeepEqual(*dep, expected_dep) {
			t.Errorf("got deployment %v, want %v", *dep, expected_dep)
		}
	})

	t.Run("should return error from repository Update", func (t *testing.T) {
		defer deployment_repository.Clear()

		deployment_repository.update_error = errors.New("update_failed")

		dep, err := deployment_service.update(mock_dep)
		if err == nil {
			t.Fatalf("got nil error, want %v", deployment_repository.update_error)
		}
		if dep != nil {
			t.Errorf("got deployment %v, want nil", dep)
		}
	})
}
