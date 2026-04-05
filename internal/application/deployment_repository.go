package application

import (
	"context"

	"github.com/salmanrf/capybara-cloud/internal/database"
)

type deployment_repository struct {
	ctx context.Context
	queries *database.Queries
}

type DeploymentRepository interface {

}

func NewDeploymentRepository(ctx context.Context, queries *database.Queries) DeploymentRepository {
	return &deployment_repository{
		ctx, queries,
	}
}