package deployment

import (
	"github.com/salmanrf/capybara-cloud/internal/database"
)

type service struct {}

type DeploymentService interface {
	CreateNewDeployment(*database.Application, *database.ApplicationConfig) (error)
}

func NewService() DeploymentService {
	return &service{}
}

func (s *service) CreateNewDeployment(app *database.Application, config *database.ApplicationConfig) (error) {
	return nil
}