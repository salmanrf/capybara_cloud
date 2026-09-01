package deployment

import (
	"context"
	"fmt"

	masbro_worker "github.com/salmanrf/capybara-cloud/internal/masbro-worker"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
)

type DeployRequest = shared_deployment.DeployRequest
type DeployStepResult = shared_deployment.DeployStepResult

type listener struct {
	ctx 					 context.Context
	in_channel 		 chan DeployRequest
	out_channel 	 chan DeployStepResult
	deploy_service Service
	masbro_service masbro_worker.Service
}

type Listener interface {
	Listen()
}

func NewListener(ctx context.Context, in_channel chan DeployRequest, out_channel chan DeployStepResult, deploy_service Service, masbro_service masbro_worker.Service) Listener {
	return &listener{
		ctx,
		in_channel,
		out_channel,
		deploy_service,
		masbro_service,
	}
}

func (l *listener) Listen() {
	go l.listenResult()

	for {
		var in DeployRequest
		var ok bool
		
		select {
		case <- l.ctx.Done():
			return
		case in, ok = <- l.in_channel:
			if !ok {
				continue
			}
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

func (l *listener) listenResult() {
	for {
		var in DeployStepResult
		var ok bool
		
		select {
		case <- l.ctx.Done():
			return
		case in, ok = <- l.out_channel:
			if !ok {
				continue
			}
		}

		dep := in.DeploymentDto

		new_status := in.DeploymentDto.Status
		switch in.DeploymentDto.Status {
		case shared_deployment.DEPLOY_STATUS_INITIATED:
			new_status = shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED
		case shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED:
			new_status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT
		case shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_BUILT:
			new_status = shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED
		case shared_deployment.DEPLOY_STATUS_BUILD_IMAGE_PUSHED:
			new_status = shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED
		case shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED:
			continue
		}

		step_err := in.DeploymentError
		if step_err != nil {
			new_status = -new_status
		}
		dep.Status = new_status

		updated, err := l.deploy_service.update(dep)
		if updated == nil|| err != nil {
			fmt.Println("Error updating deployment", err)
			continue
		}
		if step_err != nil {
			continue
		}
		if updated.Status == shared_deployment.DEPLOY_STATUS_BUILD_INSTANCE_STARTED {
			continue
		}

		dep = *updated
		req := DeployRequest{
			ApplicationDto: in.ApplicationDto,
			DeploymentDto: dep,
			ApplicationConfig: in.ApplicationConfig,
		}

		l.in_channel <- req
	}
}

// TODO: Refactor the semantic so each step progresses the status by one step
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
	result, _ := l.deploy_service.Start(dto)
	out <- result
}
