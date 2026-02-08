package application

import (
	"context"

	"github.com/salmanrf/capybara-cloud/internal/database"
)

type repository struct {
	ctx context.Context
	queries *database.Queries
}

type ApplicationRepository interface {
	FindOneWithProjectMember(database.FindOneApplicationWithProjectMemberParams) (*database.FindOneApplicationWithProjectMemberRow, error)
	CreateConfig(database.CreateApplicationConfigParams) (*database.ApplicationConfig, error)
	CreateApplication(database.CreateApplicationParams) (*database.Application, error)
	UpdateOneApplication(database.UpdateOneApplicationParams) (*database.Application, error)
}

func NewRepository(ctx context.Context, queries *database.Queries) ApplicationRepository {
	return &repository{
		ctx: ctx,
		queries: queries,
	}
}

func (r *repository) FindOneWithProjectMember(params database.FindOneApplicationWithProjectMemberParams) (*database.FindOneApplicationWithProjectMemberRow, error) {
	app_with_pm, err := r.queries.FindOneApplicationWithProjectMember(
		r.ctx,
		params,
	)

	return &app_with_pm, err
}

func (r *repository) CreateConfig(params database.CreateApplicationConfigParams) (*database.ApplicationConfig, error) {
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