package deployment

import (
	masbro_worker "github.com/salmanrf/capybara-cloud/internal/masbro-worker"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
)

type DeployRequest = shared_deployment.DeployRequest

type DeployStepResult = shared_deployment.DeployStepResult

type listener struct {
	in_channel 		 chan DeployRequest
	out_channel 	 chan DeployStepResult
	deploy_service Service
	masbro_service masbro_worker.Service
}

type Listener interface {
	Listen()
}

func NewListener(in_channel chan DeployRequest, out_channel chan DeployStepResult, deploy_service Service, masbro_service masbro_worker.Service) Listener {
	return &listener{
		in_channel,
		out_channel,
		deploy_service,
		masbro_service,
	}
}

func (l *listener) Listen() {
	for {
		in, ok := <- l.in_channel
		if !ok {
			continue
		}
		switch in.DeploymentDto.Status {
		case shared_deployment.DEPLOY_STATUS_INITIATED:
			go l.handleExtract(in, l.out_channel)
		case shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED:
			go l.handleBuild(in, l.out_channel)
		case shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT:
			go l.handlePush(in, l.out_channel)
		case shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED:
			go l.handleStart(in, l.out_channel)
		}
	}
}

func (l *listener) handleExtract(dto DeployRequest, out chan <- DeployStepResult) {
	result, _ := l.deploy_service.Extract(dto)
	out <- result
}

func (l *listener) handleBuild(dto DeployRequest, out chan <- DeployStepResult) {
	result, _ := l.deploy_service.Build(dto)
	out <- result
}

func (l *listener) handlePush(dto DeployRequest, out chan <- DeployStepResult) {
	result, _ := l.deploy_service.Push(dto)
	out <- result
}

func (l *listener) handleStart(dto DeployRequest, out chan <- DeployStepResult) {
	ins, err := l.deploy_service.createInstance(dto)
	if err != nil || ins == nil {
		return
	}

	l.masbro_service.Start(dto, *ins)
	l.deploy_service.updateInstance(ins)

	dep := &dto.DeploymentDto
	dep.Status = shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED
	
	res := shared_deployment.DeployStepResult{
		DeploymentDto: dep,
		DeploymentError: nil,
	}
	
	out <- res
}
