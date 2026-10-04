package application

import (
	"github.com/jackc/pgx/v5/pgtype"
	shared_deployment "github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
)

const (
	APP_STATUS_NOT_DEPLOYED = "not_deployed"
	APP_STATUS_RUNNING = "running"
	APP_STATUS_STOPPED = "stopped"
)

func DeriveAppStatus(instance_status pgtype.Int4) string {
	if !instance_status.Valid {
		return APP_STATUS_NOT_DEPLOYED
	}
	if instance_status.Int32 == shared_deployment.DEPLOY_INSTANCE_STATUS_RUNNING {
		return APP_STATUS_RUNNING
	}

	return APP_STATUS_STOPPED
}
