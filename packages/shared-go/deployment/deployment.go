package deployment

import (
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type DeployRequest struct {
	ApplicationDto    database.Application
	ApplicationConfig database.ApplicationConfig
	DeploymentDto     database.ApplicationDeployment
}

type DeployStepResult struct {
	ApplicationDto    database.Application
	ApplicationConfig database.ApplicationConfig
	DeploymentDto     database.ApplicationDeployment
	DeploymentError   error
}

type DeployRunResult struct {
	Instance *database.DeploymentInstance
	Error error
}
