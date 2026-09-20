package deployment

import (
	"context"
	"log/slog"

	masbro_worker "github.com/salmanrf/capybara-cloud/apps/backend/internal/masbro-worker"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
)

type DeployRequest = shared_deployment.DeployRequest
type DeployStepResult = shared_deployment.DeployStepResult

type listener struct {
	ctx 					 context.Context
	logger 				 *slog.Logger
	in_channel 		 chan DeployRequest
	out_channel 	 chan DeployStepResult
	deploy_service Service
	masbro_service masbro_worker.Service
}

type Listener interface {
	Listen()
}

func NewListener(ctx context.Context, logger *slog.Logger, in_channel chan DeployRequest, out_channel chan DeployStepResult, deploy_service Service, masbro_service masbro_worker.Service) Listener {
	return &listener{
		ctx,
		logger,
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

		app := in.ApplicationDto
		dep := in.DeploymentDto

		l.logger.Debug("Received deployment request", "app_id", app.AppID, "deployment_id", dep.AppDpID, "version", dep.VersionNumber, "status", dep.Status)

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

		app := in.ApplicationDto
		dep := in.DeploymentDto

		l.logger.Debug("Received deployment result", "app_id", app.AppID, "deployment_id", dep.AppDpID, "version", dep.VersionNumber, "status", dep.Status)

		if dep.Status < 0 {
			l.logger.Debug("Deployment status is already marked as error, skipping", "app_id", app.AppID, "deployment_id", dep.AppDpID, "version", dep.VersionNumber, "status", dep.Status)
			continue
		}

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
		default:
			l.logger.Error("Unrecognized deployment status", "app_id", app.AppID, "deployment_id", dep.AppDpID, "version", dep.VersionNumber, "status", dep.Status)
			continue
		}

		step_err := in.DeploymentError
		if step_err != nil {
			l.logger.Error("Got deployment step error", "app_id", app.AppID, "deployment_id", dep.AppDpID, "version", dep.VersionNumber, "status", dep.Status, "error", step_err.Error())
			new_status = -new_status
		}
		dep.Status = new_status

		updated, err := l.deploy_service.update(dep)
		if err != nil {
			l.logger.Error("Unable to update deployment", "app_id", app.AppID, "deployment_id", dep.AppDpID, "version", dep.VersionNumber, "status", dep.Status, "error", err.Error())
			continue
		}
		if step_err != nil || updated == nil {
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
