package deployment

import (
	"context"

	"github.com/salmanrf/capybara-cloud/internal/database"
)

type deployment_repository struct {
	ctx context.Context
	queries *database.Queries
}

type DeploymentRepository interface {
	Create(database.CreateApplicationDeploymentParams) (*database.ApplicationDeployment, error)
}

func NewDeploymentRepository(ctx context.Context, queries *database.Queries) DeploymentRepository {
	return &deployment_repository{
		ctx, queries,
	}
}

func (r *deployment_repository) Create(params database.CreateApplicationDeploymentParams) (*database.ApplicationDeployment, error) {
	row, err := r.queries.CreateApplicationDeployment(r.ctx, params)
	return &row, err
} 