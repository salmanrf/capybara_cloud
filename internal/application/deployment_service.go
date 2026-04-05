package application

type deployment_service struct {
	app_service Service
	app_deployment_repository DeploymentRepository
}	

type DeploymentService interface {
	// Deploy(application_id string) error
}

func NewDeploymentService(app_service Service, app_deployment_repository DeploymentRepository) DeploymentService {
	return &deployment_service{
		app_service, 
		app_deployment_repository,
	}
}

