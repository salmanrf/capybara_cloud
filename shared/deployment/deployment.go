package deployment

import (
	"github.com/salmanrf/capybara-cloud/internal/database"
)

type DeployRequest struct {
	ApplicationDto    database.Application
	ApplicationConfig database.ApplicationConfig
	DeploymentDto     database.ApplicationDeployment
}

type DeployStepResult struct {
	DeploymentDto   *database.ApplicationDeployment
	DeploymentError error
}
