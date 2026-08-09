package masbro_worker

import (
	"encoding/json"

	"github.com/salmanrf/capybara-cloud/internal/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
	docker "github.com/salmanrf/capybara-cloud/shared/docker"
)

type service struct {
	docker docker.Docker
}

type Service interface {
	Start(dto shared_deployment.DeployRequest, instance database.DeploymentInstance) (database.DeploymentInstance, error)
}

func New(d docker.Docker) Service {
	return &service{d}
}

func (s *service) Start(dto shared_deployment.DeployRequest, instance database.DeploymentInstance) (database.DeploymentInstance, error) {
	dep := dto.DeploymentDto
	cfg := dto.ApplicationConfig

	var env_map map[string]any
	err := json.Unmarshal(cfg.VariablesJson, &env_map); if err != nil {
		return database.DeploymentInstance{}, err
	}

	run_dto := docker.DockerRunDto{
		ImageName: dep.ContainerImgName,
		ContainerPort: int(instance.ContainerPort),
		ContainerName: instance.ContainerName,
		EnvVars: env_map,
	}

	res, err := s.docker.Run(run_dto)
	if err != nil {
		return database.DeploymentInstance{}, err
	}

	instance.HostPort = int32(res.HostPort)

	return instance, nil
}