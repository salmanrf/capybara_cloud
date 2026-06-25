package deployment

import "github.com/salmanrf/capybara-cloud/internal/database"

type DeployRequest struct {
	ApplicationDto    database.Application
	ApplicationConfig database.ApplicationConfig
	DeploymentDto     database.ApplicationDeployment
}

type DeployStepResult struct {
	DeploymentDto   *database.ApplicationDeployment
	DeploymentError error
}

type listener struct {
	in_channel 		 chan DeployRequest
	out_channel 	 chan DeployStepResult
	deploy_service Service
}

type Listener interface {
	Listen()
	handleExtract(dto DeployRequest, out chan <- DeployStepResult)
}

func NewListener(in_channel chan DeployRequest, out_channel chan DeployStepResult, deploy_service Service) Listener {
	return &listener{
		in_channel,
		out_channel,
		deploy_service,
	}
}

func (l *listener) Listen() {
	for {
		in, ok := <- l.in_channel
		if ok {
			go l.handleExtract(in, l.out_channel)
		}
	}
}

func (l *listener) handleExtract(dto DeployRequest, out chan <- DeployStepResult) {
	result, _ := l.deploy_service.Extract(dto)
	out <- result
}
