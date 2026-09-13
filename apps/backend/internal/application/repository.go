package application

import (
	"context"

	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type repository struct {
	ctx context.Context
	queries *database.Queries
}

type ApplicationRepository interface {
	FindOneComplete(database.FindOneApplicationCompleteParams) (*database.FindOneApplicationCompleteRow, error)
	UpsertConfig(database.CreateApplicationConfigParams) (*database.ApplicationConfig, error)
	CreateApplication(database.CreateApplicationParams) (*database.Application, error)
	UpdateOneApplication(database.UpdateOneApplicationParams) (*database.Application, error)
}

func NewRepository(ctx context.Context, queries *database.Queries) ApplicationRepository {
	return &repository{
		ctx: ctx,
		queries: queries,
	}
}

func (r *repository) FindOneComplete(params database.FindOneApplicationCompleteParams) (*database.FindOneApplicationCompleteRow, error) {
	app_with_pm, err := r.queries.FindOneApplicationComplete(
		r.ctx,
		params,
	)

	return &app_with_pm, err
}

func (r *repository) UpsertConfig(params database.CreateApplicationConfigParams) (*database.ApplicationConfig, error) {
	app_cfg, err := r.queries.CreateApplicationConfig(
		r.ctx,
		params,
	)

	return &app_cfg, err
}

func (r *repository) CreateApplication(params database.CreateApplicationParams) (*database.Application, error) {
	app, err := r.queries.CreateApplication(
		r.ctx,
		params,
	)

	return &app, err
}

func (r *repository) UpdateOneApplication(params database.UpdateOneApplicationParams) (*database.Application, error) {
	app, err := r.queries.UpdateOneApplication(
		r.ctx,
		params,
	)

	return &app, err
}