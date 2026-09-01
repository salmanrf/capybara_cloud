package deployment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/database"
	masbro_worker "github.com/salmanrf/capybara-cloud/internal/masbro-worker"
	config "github.com/salmanrf/capybara-cloud/pkg/utils"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
	"github.com/salmanrf/capybara-cloud/shared/docker"
	"github.com/salmanrf/capybara-cloud/shared/utils"
)

type service struct {
	ctx context.Context
	docker docker.Docker
	app_service application.Service
	masbro_service masbro_worker.Service
	port_service shared_deployment.PortAllocatorService
	deployment_repository DeploymentRepository
	deploy_chan chan DeployRequest
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
	docker_service docker.Docker,
	app_service application.Service,
	masbro_service masbro_worker.Service,
	port_service shared_deployment.PortAllocatorService,
	deployment_repository DeploymentRepository,
	deploy_chan chan DeployRequest,
) Service {
	cfg := config.GetConfig()
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
	}
}

func (s *service) update(dep database.ApplicationDeployment) (*database.ApplicationDeployment, error) {
	return nil, nil
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
	config := config.GetConfig()
	
	app, err := s.app_service.FindOne(app_id, user_id)
	if err != nil {
		return nil, err
	}
	if app == nil || !app.ApplicationConfig.AppCfgID.Valid {
		return nil, errors.New("not_found")
	}

	artifact_path, err := saveDeployArtifacts(config, *app, bundle_file_headers, bundle_file)
	if err != nil {
		return nil, err
	}

	next_version := int32(1)
	current_dp, err := s.deployment_repository.FindCurrent(app_id)
	if err != nil {
		return nil, err
	}
	if current_dp != nil {
		next_version = current_dp.VersionNumber + 1
	}

	create_dp_params := database.CreateApplicationDeploymentParams{
		AppID: app.AppID,
		ArtifactsPath: artifact_path,
		BuildPath: "abcd",
		VariablesSnapshotJson: app.ApplicationConfig.VariablesJson,
		VersionNumber: int32(next_version),
		StorageService: "localfs",
		Status: shared_deployment.DEPLOY_STATUS_INITIATED,
	}

	app_deployment, err := s.deployment_repository.Create(create_dp_params)
	if err != nil {
		return nil, err
	}

	app_dto := database.Application{
		AppID:     app.AppID,
		ProjectID: app.PmProjectID,
		Type:      app.Type,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}

	app_config_dto := database.ApplicationConfig{
		AppCfgID:      app.ApplicationConfig.AppCfgID,
		AppID:         app.AppID,
		VariablesJson: app.ApplicationConfig.VariablesJson,
		CreatedAt:     app.ApplicationConfig.CreatedAt,
		UpdatedAt:     app.ApplicationConfig.UpdatedAt,
	}

	deploy_req := DeployRequest{
		ApplicationDto:    app_dto,
		ApplicationConfig: app_config_dto,
		DeploymentDto:     *app_deployment,
	}

	s.deploy_chan <- deploy_req

	return app_deployment, nil
}

func (s *service) createInstance(dto DeployRequest) (*database.DeploymentInstance, error) {
	app := database.FindOneApplicationWithProjectMemberRow{
		Type: dto.ApplicationDto.Type,
		Name: dto.ApplicationDto.Name,
	}
	container_name := getContainerName(app)

	create_params := database.CreateDeploymentInstanceParams{
		AppID: dto.ApplicationDto.AppID,
		DeploymentID: dto.DeploymentDto.AppDpID,
		ContainerName: container_name,
		ContainerPort: dto.ApplicationConfig.Port,
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
	dep.Status = -1 * shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED

	cfg := config.GetConfig()

	build_path := getBuildDirPath(cfg, "localfs", app.Name)
	err = utils.EnsureDirExists(build_path)
	if err != nil {
		fmt.Println("Unable to create build directory")
		return res, err
	}

	tar, err := exec.LookPath("tar")
	if err != nil {
		fmt.Println("Unable to find tar executable")
		return res, err
	}
	cmd := exec.Command(tar, "-xzvf", dep.ArtifactsPath, "-C", build_path)
	extract_out := bytes.NewBuffer([]byte{})
	cmd.Stdout = extract_out
	if err := cmd.Run(); err != nil {
		fmt.Println(extract_out)
		fmt.Println("Error extracting bundle file", err)
		return res, err
	}

	dep.Status = shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED

	return res, nil
}

func (s *service) Build(dto DeployRequest) (res DeployStepResult, err error) {
	cfg := config.GetConfig()

	dep := dto.DeploymentDto
	res.ApplicationDto = dto.ApplicationDto
	res.ApplicationConfig = dto.ApplicationConfig
	defer func() { res.DeploymentDto = dep }()
	if _, err := os.ReadDir(dep.BuildPath); err != nil {
		dep.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to locate build directory")
		return res, res.DeploymentError
	}

	defer func () {
		err := os.RemoveAll(dep.BuildPath)
		if err != nil {
			fmt.Println("[Build] Cleanup: unexpected error", err)
		}
	}()

	wd, _ := os.Getwd()
	dockerfile_template_path := path.Join(wd, "templates", "Dockerfile")
	template_dockerfile, err := os.OpenFile(dockerfile_template_path, os.O_RDONLY, 0)
	if err != nil {
		dep.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to locate template Dockerfile")
		return res, res.DeploymentError
	}

	dockerfile_path := path.Join(dep.BuildPath, "Dockerfile")
	dockerfile, err := os.OpenFile(dockerfile_path, os.O_CREATE | os.O_WRONLY, 0o774)
	if err != nil {
		dep.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to locate build directory")
		return res, res.DeploymentError
	}

	_, err = io.Copy(dockerfile, template_dockerfile)
	if err != nil {
		dep.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to copy Dockerfile")
		return res, res.DeploymentError
	}

	docker_full_image_reference := getContainerImageName(dto.ApplicationDto, dto.DeploymentDto)

	docker, err := exec.LookPath("docker")
	if err != nil {
		dep.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to locate docker executable")
		return res, res.DeploymentError
	}

	docker_build_ctx_path := dep.BuildPath
	docker_build_cmd := exec.Command(docker, "build", docker_build_ctx_path, "-t", docker_full_image_reference)

	stdout, _ := docker_build_cmd.StderrPipe()
	go func () {
		temp := make([]byte, 1024)
		n, err := stdout.Read(temp)
		for ; err == nil; {
			fmt.Println(string(temp[:n]))
			n, err = stdout.Read(temp)
		}
	}()

	fmt.Printf("Starting docker build in directory: %s\n", dep.BuildPath)
	fmt.Printf("Running: %s ...\n", docker_build_cmd.String())
	err = docker_build_cmd.Run()
	if err != nil {
		dep.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to build docker image")
		return res, res.DeploymentError
	}

	dep.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
	dep.ContainerImgName = docker_full_image_reference
	dep.ContainerRegistry = cfg.DOCKER_REGISTRY

	return res, nil
}

func (s *service) Push(dto DeployRequest) (res DeployStepResult, err error) {
	new_deployment := dto.DeploymentDto
	res.ApplicationDto = dto.ApplicationDto
	res.ApplicationConfig = dto.ApplicationConfig
	defer func() { res.DeploymentDto = new_deployment }()

	repotag := utils.GetDockerRepoTagFromFullName(dto.DeploymentDto.ContainerImgName)

	dockerimg, err := s.docker.FindOneImageByName(repotag)
	if err != nil || dockerimg == nil {
		new_deployment.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		return res, errors.New("image_not_found")
	}

	err = s.docker.Push(repotag)
	if err != nil {
		err_msg := fmt.Sprintf("image_push_failed: %s", err.Error())
		new_deployment.Status = -shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		return res, errors.New(err_msg)
	}

	new_deployment.Status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED

	return res, err
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
		ContainerName: getContainerName(
			database.FindOneApplicationWithProjectMemberRow{
				Name: app.Name,
				Type: app.Type,
			},
		),
		Status: shared_deployment.DEPLOY_INSTANCE_STATUS_STOPPED,
	}
	ins, err := s.deployment_repository.CreateInstance(create_ins_params)
	if err != nil {
		res.DeploymentError = err
		return res, nil
	}

	start_res, err := s.masbro_service.Start(dto, *ins)
	if err != nil {
		res.DeploymentError = err
		return res, nil
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
	res.DeploymentError = err

	return res, nil
}

func getContainerImageName(app database.Application, dep database.ApplicationDeployment) string {
	cfg := config.GetConfig()

	regspace := fmt.Sprintf("docker.io/%s", cfg.DOCKER_REGISTRY)
	repo := utils.Slugify(app.Name)
	tag := fmt.Sprintf(
		":%s-%03s", 
		utils.DockerSafeDateString(dep.CreatedAt.Time),
		fmt.Sprintf("%d", dep.VersionNumber),
	)

	return fmt.Sprintf("%s/%s%s", regspace, repo, tag)
}

func getContainerName(app database.FindOneApplicationWithProjectMemberRow) string {
	now := time.Now()
	dateformatted := now.Format((time.DateOnly))
	timeformatted := strings.ReplaceAll(now.Format(time.TimeOnly), ":", "-")
	deploy_datestr := fmt.Sprintf("%s-%s", dateformatted, timeformatted)
	appnameslug := utils.Slugify(app.Name)

	full_container_name := fmt.Sprintf("%s-%s-%s", app.Type, appnameslug, deploy_datestr)

	return full_container_name
}

func getBuildDirPath(config config.Config, storage_service string, app_name string) string {
	_ = storage_service

	now := time.Now()
	dateformatted := now.Format(time.DateOnly)
	timeformatted := strings.ReplaceAll(now.Format(time.TimeOnly), ":", "-")
	deploy_datestr := fmt.Sprintf("%s-%s", dateformatted, timeformatted)
	appnameslug := utils.Slugify(app_name)

	full_path := fmt.Sprintf("%s/build-%s-%s", config.BASE_BUILD_PATH, appnameslug, deploy_datestr) 
	
	return full_path
}

func getArtifactDirPath(config config.Config, storage_service string, app database.FindOneApplicationWithProjectMemberRow) string {
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
	config config.Config,
	app database.FindOneApplicationWithProjectMemberRow, 
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
		fmt.Println("Error writing bundle file", full_path, err)
		return "", err
	} 

	return full_path, nil
}

