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
	"github.com/salmanrf/capybara-cloud/pkg/utils"
	"github.com/salmanrf/capybara-cloud/shared/docker"
)

type service struct {
	ctx context.Context
	docker docker.Docker
	app_service application.Service
	port_service PortAllocatorService
	deployment_repository DeploymentRepository
	deploy_chan chan DeployRequest
}	

type Service interface {
	Deploy(user_id string, app_id string, bundle_file_headers *multipart.FileHeader, bundle multipart.File) (*database.ApplicationDeployment, error)
	Extract(dto DeployRequest) (DeployStepResult, error)
	Build(dto DeployRequest) (DeployStepResult, error)
	Push(dto DeployRequest) (DeployStepResult, error)
}

func NewService(
	ctx context.Context,
	docker_service docker.Docker,
	app_service application.Service,
	port_service PortAllocatorService,
	deployment_repository DeploymentRepository,
	deploy_chan chan DeployRequest,
) Service {
	cfg := utils.GetConfig()
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
		port_service,
		deployment_repository,
		deploy_chan,
	}
}

func (s *service) Extract(dto DeployRequest) (DeployStepResult, error) {
	app := dto.ApplicationDto
	dep := dto.DeploymentDto

	res := DeployStepResult{
		DeploymentDto: &dep,
		DeploymentError: nil,
	}
	dep.Status = -1 * DEPLOY_STATUS_BUILD_EXTRACTED

	cfg := utils.GetConfig()

	build_path := getBuildDirPath(cfg, "localfs", app.Name)
	err := utils.EnsureDirExists(build_path)
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

	res.DeploymentDto.Status = DEPLOY_STATUS_BUILD_EXTRACTED

	return res, nil
}

func (s *service) Build(dto DeployRequest) (res DeployStepResult, err error) {
	cfg := utils.GetConfig()
	
	dep := dto.DeploymentDto
	if _, err := os.ReadDir(dep.BuildPath); err != nil {
		dto.DeploymentDto.Status = -DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentDto = &dto.DeploymentDto
		res.DeploymentError = errors.New("Unable to locate build directory")
		return res, res.DeploymentError
	}

	res.DeploymentDto = &dto.DeploymentDto
	
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
		res.DeploymentDto.Status = -DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to locate template Dockerfile")
		return res, res.DeploymentError
	}

	dockerfile_path := path.Join(dep.BuildPath, "Dockerfile")
	dockerfile, err := os.OpenFile(dockerfile_path, os.O_CREATE | os.O_WRONLY, 0o774)
	if err != nil {
		res.DeploymentDto.Status = -DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to locate build directory")
		return res, res.DeploymentError
	}

	_, err = io.Copy(dockerfile, template_dockerfile)
	if err != nil {
		res.DeploymentDto.Status = -DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to copy Dockerfile")
		return res, res.DeploymentError
	}

	docker_full_image_reference := getContainerImageName(dto.ApplicationDto, dto.DeploymentDto)

	docker, err := exec.LookPath("docker")
	if err != nil {
		res.DeploymentDto.Status = -DEPLOY_STATUS_BUILD_IMAGE_BUILT
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
		res.DeploymentDto.Status = -DEPLOY_STATUS_BUILD_IMAGE_BUILT
		res.DeploymentError = errors.New("Unable to build docker image")
		return res, res.DeploymentError
	}

	res.DeploymentDto.Status = DEPLOY_STATUS_BUILD_IMAGE_BUILT
	res.DeploymentDto.ContainerImgName = docker_full_image_reference
	res.DeploymentDto.ContainerRegistry = cfg.DOCKER_REGISTRY

	return res, err
}

func (s *service) Push(dto DeployRequest) (res DeployStepResult, err error) {
	new_deployment := dto.DeploymentDto
	res.DeploymentDto = &new_deployment
	
	dockerimg, err := s.docker.FindOneImageByName(dto.DeploymentDto.ContainerImgName)
	if err != nil || dockerimg == nil {
		new_deployment.Status = -DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		return res, errors.New("image_not_found")
	}

	err = s.docker.Push(dockerimg)
	if err != nil {
		err_msg := fmt.Sprintf("image_push_failed: %s", err.Error())
		new_deployment.Status = -DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		return res, errors.New(err_msg)
	}

	new_deployment.Status = DEPLOY_STATUS_BUILD_IMAGE_PUSHED

	return res, err
}

func (s *service) handleDeployRequest(dto DeployRequest) {
	app := dto.ApplicationDto
	dep := dto.DeploymentDto
	cfg := dto.ApplicationConfig

	appDetails := database.FindOneApplicationWithProjectMemberRow{
		Name: app.Name,
		Type: app.Type,
	}
	containerName := getContainerName(appDetails)
	
	host_port, err := s.port_service.GetFreePort()
	if err != nil {
		fmt.Println("handleDeployRequest error allocating port", err)
	}
	create_ins_params := database.CreateDeploymentInstanceParams{
		AppID: app.AppID,
		DeploymentID: dep.AppDpID,
		ContainerName: containerName,
		ContainerPort: cfg.Port,
		HostPort: int32(host_port),
	}
	s.deployment_repository.CreateInstance(create_ins_params)
}

func (s *service) Deploy(user_id string, app_id string, bundle_file_headers *multipart.FileHeader, bundle_file multipart.File) (*database.ApplicationDeployment, error) {
	config := utils.GetConfig()
	
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
		Status: DEPLOY_STATUS_INITIATED,
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

func getContainerImageName(app database.Application, dep database.ApplicationDeployment) string {
	cfg := utils.GetConfig()
	
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

	full_container_name := fmt.Sprintf("%s-%s-%s", app.Type, appnameslug, deploy_datestr )

	return full_container_name
}

func getBuildDirPath(config utils.Config, storage_service string, app_name string) string {
	_ = storage_service
	
	now := time.Now()
	dateformatted := now.Format(time.DateOnly)
	timeformatted := strings.ReplaceAll(now.Format(time.TimeOnly), ":", "-")
	deploy_datestr := fmt.Sprintf("%s-%s", dateformatted, timeformatted)
	appnameslug := utils.Slugify(app_name)

	full_path := fmt.Sprintf("%s/build-%s-%s", config.BASE_BUILD_PATH, appnameslug, deploy_datestr) 
	
	return full_path
}

func getArtifactDirPath(config utils.Config, storage_service string, app database.FindOneApplicationWithProjectMemberRow) string {
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
	config utils.Config,
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

func startContainerizedApp(
	globalcfg utils.Config,
	app database.FindOneApplicationWithProjectMemberRow,
	appdp database.ApplicationDeployment,
) error {
	artifact_path := appdp.ArtifactsPath
	pathsegs := strings.Split(artifact_path, "/")
	filename := pathsegs[len(pathsegs) - 1]
	artifact_dir := strings.Split(artifact_path, filename)[0]

	_, err := os.OpenFile(artifact_path, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("Unable to open sample deployment bundle", err)
		return err
	}
	
	tar, err := exec.LookPath("tar")
	if err != nil {
		fmt.Println("Unable to find 'tar' executable", err)
		return err
	}

	cmd := exec.Command(tar, "-xzvf", artifact_path, "-C", artifact_dir)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Unable to find 'tar' executable", err)
		return err
	}

	wd, _ := os.Getwd()
	dockerfileTemplatePath := path.Join(wd, "templates", "Dockerfile") 
	dockerfileTemplate, err := os.OpenFile(dockerfileTemplatePath, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("Unable to open Dockerfile template", err)
		return err
	}

	targetDockerfilePath := path.Join(artifact_dir, "Dockerfile")
	targetDockerfile, err := os.OpenFile(targetDockerfilePath, os.O_CREATE | os.O_WRONLY, 0o774)
	if err != nil {
		fmt.Println("Unable to open Dockerfile copy target", err)
		return err
	}

	_, err = io.Copy(targetDockerfile, dockerfileTemplate)
	if err != nil {
		fmt.Println("Unable to copy Dockerfile template", err)
		return err
	}

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Unable to find docker executables: ", err)
		return err
	}

	buildLogs := bytes.NewBuffer([]byte{})
	buildErrs := bytes.NewBuffer([]byte{})

	containerName := getContainerName(app)
	dockerBuildCtxDir := artifact_dir
	dockerRegNamespace := globalcfg.DOCKER_REGISTRY
	dockerRegRepository := utils.Slugify(app.Name)
	dockerTag := strings.Split(
		containerName,
		fmt.Sprintf("%s-%s-", app.Type, dockerRegRepository),
	)[1]
	dockerFullTag := fmt.Sprintf("%s/%s:%s", dockerRegNamespace, dockerRegRepository, dockerTag)
	dockerBuildCmd := exec.Command(dockerPath, "build", "-t", dockerFullTag, dockerBuildCtxDir)
	dockerBuildCmd.Stdout = buildLogs

	err = dockerBuildCmd.Run()
	if err != nil {
		fmt.Println("BUILD LOGS", buildLogs)
		fmt.Println("BUILD ERRS", buildErrs)
		fmt.Println("Unable to build container image " + dockerFullTag, err)
		return err
	}

	runLogs := bytes.NewBuffer([]byte{})
	runErrs := bytes.NewBuffer([]byte{})
	dockerRunCmd := exec.Command(dockerPath, "run", "--name", containerName, "-d", "-p", "8080:3000", dockerFullTag)
	dockerRunCmd.Stdout = runLogs
	err = dockerRunCmd.Run()
	if err != nil {
		fmt.Println("RUN LOGS", runLogs)
		fmt.Println("RUN ERRS", runErrs)
		fmt.Println("Unable to start container " + containerName, err)
		return err
	}

	return nil
}

