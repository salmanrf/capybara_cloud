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

	"github.com/moby/moby/client"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

type service struct {
	ctx context.Context
	docker *client.Client
	app_service application.Service
	deployment_repository DeploymentRepository
}	

type Service interface {
	Deploy(user_id string, app_id string, bundle_file_headers *multipart.FileHeader, bundle multipart.File) (*database.ApplicationDeployment, error)
	// FindCurrentDeployment(app_id string) (*database.ApplicationDeployment, error)
}

func NewService(ctx context.Context, app_service application.Service, deployment_repository DeploymentRepository) Service {
	docker, _ := client.New(client.FromEnv)
	
	return &service{
		ctx,
		docker,
		app_service, 
		deployment_repository,
	}
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

	now := time.Now()
	deploy_datestr := fmt.Sprintf("%s-%s", now.Format(time.DateOnly), now.Format(time.TimeOnly))
	artifact_path, err := saveDeployArtifacts(config, *app, bundle_file_headers, bundle_file)
	container_name := fmt.Sprintf("%s:%s", deploy_datestr, utils.Slugify(app.Name)) 

	create_dp_params := database.CreateApplicationDeploymentParams{
		AppID: app.AppID,
		ArtifactsPath: artifact_path,
		ProcessName: "",
		ContainerName: container_name,
		VariablesSnapshotJson: app.ApplicationConfig.VariablesJson,
	}

	app_deployment, err := s.deployment_repository.Create(create_dp_params)
	if err != nil {
		return nil, err
	}

	return app_deployment, nil
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

func getArtifactDirPath(config utils.Config, storage_service string, app database.FindOneApplicationWithProjectMemberRow) string {
	_ = storage_service
	
	now := time.Now()
	dateformatted := now.Format(time.DateOnly)
	timeformatted := strings.ReplaceAll(now.Format(time.TimeOnly), ":", "-")
	deploy_datestr := fmt.Sprintf("%s-%s", dateformatted, timeformatted)
	appnameslug := utils.Slugify(app.Name)

	full_path := fmt.Sprintf("%s/artifact-%s-%s", config.BASE_TEMP_PATH, appnameslug, deploy_datestr) 
	
	return full_path
}

// TODO: Fix parallel tests failure due to contention in deleting /tmp/capybara-cloud
func saveDeployArtifacts(
	config utils.Config,
	app database.FindOneApplicationWithProjectMemberRow, 
	fileheaders *multipart.FileHeader,
	bundle_file multipart.File,
) (string, error) { 
	dir := getArtifactDirPath(config, "", app)
	full_path := dir + "/" + fileheaders.Filename

	_, err := os.Stat(config.BASE_TEMP_PATH)
	if err != nil {
		errmsg := err.Error()
		fmt.Printf("Deploy warning: Unable to check root artifacts directory: %v\n", err.Error())
		if strings.Contains(errmsg, "no such file") {
			err = os.Mkdir(config.BASE_TEMP_PATH, 0o774)
			if err != nil {
				fmt.Printf("Deploy warning: Unable to create root artifacts directory: %v\n", err.Error())
				return "", errors.New("Internal error")
			}	else {
				fmt.Println("Deploy info: Root artifacts dir created!")
			}	
		} else {
			return "", errors.New("Internal error")
		}
	}
	err = os.Mkdir(dir, 0o774)
	if err != nil {
		fmt.Printf("Deploy error: Unable to create artifact directory: %v\n", err.Error())
		return "", errors.New("Internal error")
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
	appcfg database.ApplicationConfig,
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

	dockerBuildCtxDir := artifact_dir
	dockerRegNamespace := globalcfg.DOCKER_REGISTRY
	dockerRegRepository := utils.Slugify(app.Name)
	dockerTag := strings.Split(
		appdp.ContainerName,
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
	dockerRunCmd := exec.Command(dockerPath, "run", "--name", appdp.ContainerName, "-d", "-p", "8080:3000", dockerFullTag)
	dockerRunCmd.Stdout = runLogs
	err = dockerRunCmd.Run()
	if err != nil {
		fmt.Println("RUN LOGS", runLogs)
		fmt.Println("RUN ERRS", runErrs)
		fmt.Println("Unable to start container " + appdp.ContainerName, err)
		return err
	}

	return nil
}

