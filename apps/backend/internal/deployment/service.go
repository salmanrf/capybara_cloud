package deployment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	pkgerr "github.com/pkg/errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/salmanrf/capybara-cloud/apps/backend/internal/application"
	masbro_worker "github.com/salmanrf/capybara-cloud/apps/backend/internal/masbro-worker"
	locutils "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/docker"
	errutils "github.com/salmanrf/capybara-cloud/packages/shared-go/errors"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

type service struct {
	ctx context.Context
	docker docker.Docker
	app_service application.Service
	masbro_service masbro_worker.Service
	port_service shared_deployment.PortAllocatorService
	deployment_repository DeploymentRepository
	deploy_chan chan DeployRequest
	logger *slog.Logger
}	

type Service interface {
	Deploy(user_id string, app_id string, bundle_file_headers *multipart.FileHeader, bundle multipart.File) (*database.ApplicationDeployment, error)
	Extract(dto DeployRequest) (DeployStepResult, error)
	Build(dto DeployRequest) (DeployStepResult, error)
	Push(dto DeployRequest) (DeployStepResult, error)
	Start(dto DeployRequest) (DeployStepResult, error)
	update(dep database.ApplicationDeployment) (*database.ApplicationDeployment, error)
	updateStatus(dep database.ApplicationDeployment) (*database.ApplicationDeployment, error)
	createInstance(dto DeployRequest) (*database.DeploymentInstance, error) 
	updateInstance(*database.DeploymentInstance) (*database.DeploymentInstance, error) 
}

func NewService(
	ctx context.Context,
	logger *slog.Logger,
	docker_service docker.Docker,
	app_service application.Service,
	masbro_service masbro_worker.Service,
	port_service shared_deployment.PortAllocatorService,
	deployment_repository DeploymentRepository,
	deploy_chan chan DeployRequest,
) Service {
	cfg := locutils.GetConfig()
	err := utils.EnsureDirExists(cfg.BASE_ARTIFACT_PATH)
	if err != nil {
		panic(err)
	}
	err = utils.EnsureDirExists(cfg.BASE_BUILD_PATH)
	if err != nil {
		panic(err)
	}

	return &service{
		ctx,
		docker_service,
		app_service,
		masbro_service,
		port_service,
		deployment_repository,
		deploy_chan,
		logger,
	}
}

func (s *service) update(dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
	params := database.UpdateDeploymentParams{
		AppDpID: dep.AppDpID,
		AppID: dep.AppID,
		ArtifactsPath: dep.ArtifactsPath,
		VariablesSnapshotJson: dep.VariablesSnapshotJson,
		StorageService: dep.StorageService,
		VersionNumber: dep.VersionNumber,
		Status: dep.Status,
		BuildPath: dep.BuildPath,
		ContainerRegistry: dep.ContainerRegistry,
		ContainerNamespace: dep.ContainerNamespace,
		ContainerRepository: dep.ContainerRepository,
		ContainerTag: dep.ContainerTag,
	}
	res, err := s.deployment_repository.Update(params)
	return res, err
}

func (s *service) updateStatus(dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
	params := database.UpdateDeploymentStatusParams{
		AppDpID: dep.AppDpID,
		Status: dep.Status,
	}
	res, err := s.deployment_repository.UpdateStatus(params)	
	return res, err
}

func (s *service) Deploy(user_id string, app_id string, bundle_file_headers *multipart.FileHeader, bundle_file multipart.File) (*database.ApplicationDeployment, error) {
	config := locutils.GetConfig()

	app, err := s.app_service.FindOneComplete(app_id, user_id)
	if err != nil {
		errmsg := errors.Join(errors.New("[Deploy] Unable to find application"), err)
		s.logger.Error(errmsg.Error(), "user_id", user_id, "app_id", app_id)
		return nil, err
	}
	if app == nil {
		s.logger.Error("[Deploy] Application not found", "user_id", user_id, "app_id", app_id)
		return nil, errors.New("not_found")
	}
	if !app.ApplicationConfig.AppCfgID.Valid {
		s.logger.Error("[Deploy] Application config not found", "user_id", user_id, "app_id", app_id)
		return nil, errors.New("config_not_found")
	}
	if !app.ProjectMember.ProjectID.Valid {
		s.logger.Error("[Deploy] Permission denied, invalid project member", "user_id", user_id, "app_id", app_id)
		return nil, errors.New("permission_denied")
	}

	artifact_path, err := saveDeployArtifacts(config, *app, bundle_file_headers, bundle_file)
	if err != nil {
		errmsg := errors.Join(errors.New("[Deploy] Unable to save deployment artifact"), err)
		s.logger.Error(errmsg.Error(), "user_id", user_id, "app_id", app_id)
		return nil, errmsg
	}

	next_version := int32(1)
	current_dp, err := s.deployment_repository.FindCurrent(app_id)
	if err != nil {
		if err.Error() != "not_found" {
			errmsg := errors.Join(errors.New("[Deploy] - Unable to find current deployment"), err)
			s.logger.Error(errmsg.Error(), "user_id", user_id, "app_id", app_id)
			return nil, errmsg
		}
	}
	if current_dp != nil {
		next_version = current_dp.VersionNumber + 1
	}

	container_tag := getContainerTag(database.ApplicationDeployment{
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		VersionNumber: next_version,
	})

	create_dp_params := database.CreateApplicationDeploymentParams{
		AppID: app.AppID,
		ArtifactsPath: artifact_path,
		BuildPath: getBuildDirPath(config, "localfs", app.Name),
		VariablesSnapshotJson: app.ApplicationConfig.VariablesJson,
		VersionNumber: int32(next_version),
		StorageService: "localfs",
		Status: shared_deployment.DEPLOY_STATUS_INITIATED,
		ContainerRegistry: pgtype.Text{String: config.DOCKER_REGISTRY, Valid: true},
		ContainerNamespace: pgtype.Text{String: config.DOCKER_NAMESPACE, Valid: true},
		ContainerRepository: pgtype.Text{String: utils.Slugify(app.Name), Valid: true},
		ContainerTag: pgtype.Text{String: container_tag, Valid: true},
	}

	app_deployment, err := s.deployment_repository.Create(create_dp_params)
	if err != nil {
		errmsg := errors.Join(errors.New("[Deploy] - Unable to create deployment"), err)
		s.logger.Error(errmsg.Error(), "user_id", user_id, "app_id", app_id, "artifact_path", artifact_path)
		return nil, errmsg
	}

	app_dto := database.Application{
		AppID:     app.AppID,
		ProjectID: app.ProjectID,
		Type:      app.Type,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}

	deploy_req := DeployRequest{
		ApplicationDto:    app_dto,
		ApplicationConfig: app.ApplicationConfig,
		DeploymentDto:     *app_deployment,
	}

	s.logger.Info(
		"[Deploy] - Deployment created successfully, passing to listener",
		"user_id", user_id,
		"app_id", app_id,
		"deployment_id", app_deployment.AppDpID,
		"artifact_path", artifact_path,
	)
	s.deploy_chan <- deploy_req

	return app_deployment, nil
}

func (s *service) Extract(dto DeployRequest) (res DeployStepResult, err error) {
	app := dto.ApplicationDto
	dep := dto.DeploymentDto

	res = DeployStepResult{
		ApplicationDto: dto.ApplicationDto,
		ApplicationConfig: dto.ApplicationConfig,
		DeploymentDto: dep,
		DeploymentError: nil,
	}
	defer func() { res.DeploymentDto = dep }()

	build_path := dep.BuildPath
	err = utils.EnsureDirExists(build_path)
	if err != nil {
		errmsg := errors.Join(
			errors.New("[Deploy Step - Extract] - Unable to create build directory"),
			err,
		)
		s.logger.Error(
			errmsg.Error(),
			"app_id", app.AppID,
			"deployment_id", dep.AppDpID,
			"build_path", build_path,
			"artifact_path", dep.ArtifactsPath,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	tar, err := exec.LookPath("tar")
	if err != nil {
		errmsg := errors.Join(
			errors.New("[Deploy Step - Extract] - Unable to find tar executable"),
			err,
		)
		s.logger.Error(
			errmsg.Error(),
			"app_id", app.AppID,
			"deployment_id", dep.AppDpID,
			"build_path", build_path,
			"artifact_path", dep.ArtifactsPath,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	cmd := exec.Command(tar, "-xzvf", dep.ArtifactsPath, "-C", build_path)
	extract_out := bytes.NewBuffer([]byte{})
	cmd.Stdout = extract_out
	cmd.Stderr = extract_out
	if err := cmd.Run(); err != nil {
		errmsg := errors.Join(
			errors.New("[Deploy Step - Extract] - Unable to extract bundle file"),
			err,
		)
		s.logger.Error(
			errmsg.Error(),
			"app_id", app.AppID,
			"deployment_id", dep.AppDpID,
			"build_path", build_path,
			"artifact_path", dep.ArtifactsPath,
			"tar_output", extract_out.String(),
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	s.logger.Info(
		"[Deploy Step - Extract] - Bundle extracted successfully",
		"app_id", app.AppID,
		"deployment_id", dep.AppDpID,
		"build_path", build_path,
		"artifact_path", dep.ArtifactsPath,
	)

	return res, nil
}

func (s *service) Build(dto DeployRequest) (res DeployStepResult, err error) {
	cfg := locutils.GetConfig()

	app := dto.ApplicationDto
	dep := dto.DeploymentDto
	res.ApplicationDto = dto.ApplicationDto
	res.ApplicationConfig = dto.ApplicationConfig
	defer func() { res.DeploymentDto = dep }()

	if !dep.ContainerTag.Valid {
		dep.ContainerRegistry = pgtype.Text{String: cfg.DOCKER_REGISTRY, Valid: true}
		dep.ContainerNamespace = pgtype.Text{String: cfg.DOCKER_NAMESPACE, Valid: true}
		dep.ContainerRepository = pgtype.Text{String: utils.Slugify(app.Name), Valid: true}
		dep.ContainerTag = pgtype.Text{String: getContainerTag(dep), Valid: true}
	}
	docker_full_image_reference := utils.FullImageRef(dep)

	if _, err := os.ReadDir(dep.BuildPath); err != nil {
		errmsg := errutils.WithAttrs(pkgerr.WithStack(errors.Join(err)))
		s.logger.Error(
			"[Deploy Step - Build] - Unable to locate build directory", 
			"error",
			errmsg,
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"build_path", dep.BuildPath,
			"image", docker_full_image_reference,
		)
		res.DeploymentError = err
		return res, res.DeploymentError
	}

	defer func () {
		err := os.RemoveAll(dep.BuildPath)
		if err != nil {
			s.logger.Error("[Deploy Step - Build] Unable to cleanup build directory", "error", err, "app_id", app.AppID, "deployment_id", dep.AppDpID, "build_path", dep.BuildPath)
		}
	}()

	dockerfile_template_path := path.Join(locutils.GetConfig().DOCKER_TEMPLATES_DIR, "Dockerfile")
	template_dockerfile, err := os.OpenFile(dockerfile_template_path, os.O_RDONLY, 0)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Build] - Unable to locate template Dockerfile",
			"error",
			errmsg, 
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"build_path", dep.BuildPath,
			"image", docker_full_image_reference,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	dockerfile_path := path.Join(dep.BuildPath, "Dockerfile")
	dockerfile, err := os.OpenFile(dockerfile_path, os.O_CREATE | os.O_WRONLY, 0o774)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Build] - Unable to read Dockerfile - Unable to read Dockerfile",
			"error",
			errmsg, 
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"build_path", dep.BuildPath,
			"image", docker_full_image_reference,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	_, err = io.Copy(dockerfile, template_dockerfile)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Build] - Unable to copy Dockerfile",
			"error",
			errmsg, 
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"build_path", dep.BuildPath,
			"image", docker_full_image_reference,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	build_params := docker.DockerBuildDto{ContextPath: dep.BuildPath, Tag: docker_full_image_reference}
	err = s.docker.Build(build_params)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Build] - Unable to build docker image",
			"error",
			errmsg, 
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"build_path", dep.BuildPath,
			"image", docker_full_image_reference,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	s.logger.Info(
		"[Deploy Step - Build] - Docker image built successfully",
		"app_id", app.AppID,
		"deployment_id", dep.AppDpID,
		"build_path", dep.BuildPath,
		"image", docker_full_image_reference,
	)

	return res, nil
}

func (s *service) Push(dto DeployRequest) (res DeployStepResult, err error) {
	app := dto.ApplicationDto
	dep := dto.DeploymentDto
	res.ApplicationDto = dto.ApplicationDto
	res.ApplicationConfig = dto.ApplicationConfig
	res.DeploymentDto = dep

	repotag := fmt.Sprintf("%s:%s", dep.ContainerRepository.String, dep.ContainerTag.String)
	image := utils.FullImageRef(dep)

	s.logger.Debug(
		"[Deploy Step - Push] Pushing image",
		"app_id", app.AppID,
		"dep_id", dep.AppDpID,
		"image", image,
	)

	dockerimg, err := s.docker.FindOneImageByRepoTag(repotag)
	if err != nil || dockerimg == nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Push] Unable to find docker image",
			"error",
			errmsg, 
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"image", image,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	err = s.docker.Push(image)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Push] Unable to push docker image",
			"error",
			errmsg, 
			"app_id", app.AppID, 
			"deployment_id", dep.AppDpID, 
			"image", image,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	s.logger.Info(
		"[Deploy Step - Push] Docker image pushed successfully",
		"app_id", app.AppID,
		"deployment_id", dep.AppDpID,
		"image", image,
	)

	return res, nil
}

func (s *service) Start(dto DeployRequest) (DeployStepResult, error) {
	app := dto.ApplicationDto
	dep := dto.DeploymentDto
	cfg := dto.ApplicationConfig
	res := DeployStepResult{
		ApplicationDto: app,
		ApplicationConfig: cfg,
		DeploymentDto: dep,
		DeploymentError: nil,
	}
	create_ins_params := database.CreateDeploymentInstanceParams{
		AppID: app.AppID,
		DeploymentID: dep.AppDpID,
		ContainerName: getContainerName(dep),
		ContainerPort: cfg.Port.Int32,
		Status: shared_deployment.DEPLOY_INSTANCE_STATUS_STOPPED,
	}

	s.logger.Debug(
		"[Deploy Step - Start] Starting Container",
		"app_id", app.AppID,
		"dep_id", dep.AppDpID,
		"container", create_ins_params.ContainerName,
		"port", create_ins_params.ContainerPort,
	)
	
	ins, err := s.deployment_repository.CreateInstance(create_ins_params)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Start] Unable to create deployment instance",
			"error",
			errmsg, 
			"app_id", app.AppID,
			"dep_id", dep.AppDpID,
			"container", create_ins_params.ContainerName,
			"port", create_ins_params.ContainerPort,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	start_res, err := s.masbro_service.Start(dto, *ins)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Start] Unable to start container",
			"error",
			errmsg, 
			"app_id", app.AppID,
			"dep_id", dep.AppDpID,
			"container", create_ins_params.ContainerName,
			"port", create_ins_params.ContainerPort,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	update_params := database.UpdateDeploymentInstanceParams{
		AppID: start_res.AppID,
		DeploymentID: start_res.DeploymentID,
		InstanceID: start_res.InstanceID,
		Status: start_res.Status,
		ContainerName: start_res.ContainerName,
		HostPort: start_res.HostPort,
		ContainerPort: start_res.ContainerPort,
	}
	ins, err = s.deployment_repository.UpdateInstance(update_params)
	if err != nil {
		errmsg := errutils.WithAttrs(
			pkgerr.WithStack(err),
		)
		s.logger.Error(
			"[Deploy Step - Start] Unable to update deplyment instance",
			"error",
			errmsg, 
			"app_id", app.AppID,
			"dep_id", dep.AppDpID,
			"container", create_ins_params.ContainerName,
			"port", create_ins_params.ContainerPort,
		)
		res.DeploymentError = errmsg
		return res, res.DeploymentError
	}

	return res, nil
}

func (s *service) createInstance(dto DeployRequest) (*database.DeploymentInstance, error) {
	container_name := getContainerName(dto.DeploymentDto)

	create_params := database.CreateDeploymentInstanceParams{
		AppID: dto.ApplicationDto.AppID,
		DeploymentID: dto.DeploymentDto.AppDpID,
		ContainerName: container_name,
		ContainerPort: dto.ApplicationConfig.Port.Int32,
	}

	instance, err := s.deployment_repository.CreateInstance(create_params)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *service) updateInstance(ins *database.DeploymentInstance) (*database.DeploymentInstance, error) {
	update_params := database.UpdateDeploymentInstanceParams{
		InstanceID: ins.InstanceID,
		AppID: ins.AppID,
		DeploymentID: ins.DeploymentID,
		ContainerName: ins.ContainerName,
		HostPort: ins.HostPort,
		ContainerPort: ins.ContainerPort,
	}

	instance, err := s.deployment_repository.UpdateInstance(update_params)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func getContainerTag(dep database.ApplicationDeployment) string {
	return fmt.Sprintf("%03d-%s", dep.VersionNumber, utils.DockerSafeDateString(dep.CreatedAt.Time))
}

func getContainerName(dep database.ApplicationDeployment) string {
	return fmt.Sprintf("%s-%s-%s", dep.ContainerRepository.String, dep.ContainerTag.String, uuid.New())
}

func getBuildDirPath(config locutils.Config, storage_service string, app_name string) string {
	_ = storage_service

	now := time.Now()
	dateformatted := now.Format(time.DateOnly)
	timeformatted := strings.ReplaceAll(now.Format(time.TimeOnly), ":", "-")
	deploy_datestr := fmt.Sprintf("%s-%s", dateformatted, timeformatted)
	appnameslug := utils.Slugify(app_name)

	full_path := fmt.Sprintf("%s/build-%s-%s", config.BASE_BUILD_PATH, appnameslug, deploy_datestr) 

	return full_path
}

func getArtifactDirPath(config locutils.Config, storage_service string, app database.FindOneApplicationCompleteRow) string {
	_ = storage_service

	now := time.Now()
	dateformatted := now.Format(time.DateOnly)
	timeformatted := strings.ReplaceAll(now.Format(time.TimeOnly), ":", "-")
	deploy_datestr := fmt.Sprintf("%s-%s", dateformatted, timeformatted)
	appnameslug := utils.Slugify(app.Name)

	full_path := fmt.Sprintf("%s/artifact-%s-%s", config.BASE_ARTIFACT_PATH, appnameslug, deploy_datestr) 

	return full_path
}

func saveDeployArtifacts(
	config locutils.Config,
	app database.FindOneApplicationCompleteRow, 
	fileheaders *multipart.FileHeader,
	bundle_file multipart.File,
) (string, error) { 
	dir := getArtifactDirPath(config, "", app)
	full_path := dir + "/" + fileheaders.Filename

	err := utils.EnsureDirExists(dir)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(full_path, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o774)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024 * 1024)
	nread, err := bundle_file.Read(buf)

	for ; nread > 0 && err == nil; {
		content := buf[0:nread]
		file.Write(content)
		nread, err = bundle_file.Read(buf)
	}
	file.Close()
	if err != io.EOF {
		return "", err
	} 

	return full_path, nil
}

