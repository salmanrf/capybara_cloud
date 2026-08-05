package masbro_worker

import (
	"github.com/salmanrf/capybara-cloud/internal/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
	"github.com/salmanrf/capybara-cloud/shared/docker"
)

type service struct {
	docker docker.Docker
}

type Service interface {
	Start(dto shared_deployment.DeployRequest) (*database.DeploymentInstance, error)
}

func New(d docker.Docker) Service {
	return &service{d}
}

func (s *service) Start(dto shared_deployment.DeployRequest) (res *database.DeploymentInstance, err error) {
	err = s.docker.Pull(dto.DeploymentDto.ContainerImgName)
	if err != nil {
		return nil, err
	}

	return nil, nil
}