package deployment

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/database"
)

type deployment_repository struct {
	ctx context.Context
	queries *database.Queries
}

type DeploymentRepository interface {
	Create(database.CreateApplicationDeploymentParams) (*database.ApplicationDeployment, error)
	FindCurrent(app_id string) (*database.ApplicationDeployment, error)
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

func (r *deployment_repository) FindCurrent(app_id string) (*database.ApplicationDeployment, error) {
	app_uuid := pgtype.UUID{}
	app_uuid.Scan(app_id)
	row, err := r.queries.FindCurrentDeployment(
		r.ctx,
		app_uuid,
	)

	if strings.Contains(err.Error(), "no rows") {
		return nil, errors.New("not_found")
	}
	
	return &row, err
} 