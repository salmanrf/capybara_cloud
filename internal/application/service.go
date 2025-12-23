package application

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
)

type Service interface {
	Create(user_id string, dto dto.CreateApplicationDto) (*database.Application, error)
}

type service struct {
	ctx context.Context
	conn *pgxpool.Pool
	queries *database.Queries
	project_service project.Service
}

func NewService(ctx context.Context, conn *pgxpool.Pool, queries *database.Queries, project_service project.Service) Service {
	return &service{
		ctx,
		conn,
		queries,
		project_service,
	}
}

func (s *service) Create(user_id string, dto dto.CreateApplicationDto) (*database.Application, error) {
	proj_user, err :=  s.project_service.FindByIdAndRole(user_id, dto.ProjectID, []string{"member"})

	if err != nil {
		return nil, err
	}

	if proj_user == nil {
		permission_err := errors.New("permisssion_denied")
		return nil, permission_err
	}

	project_uuid := pgtype.UUID{}
	project_uuid.Scan(dto.ProjectID)
	
	dbparams := database.CreateApplicationParams{
		Type: dto.Type,
		ProjectID: project_uuid,
		Name: dto.Name,
	}

	new_application, err := s.queries.CreateApplication(s.ctx, dbparams)

	if err != nil {
		return nil, err
	}
	
	return &new_application, nil
}