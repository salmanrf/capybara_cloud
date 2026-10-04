package project

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/organization"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/user"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type Service interface {
	Create(user_id string, org_id string, project_name string) (*database.Project, error)
	UpdateOne(dto *database.FindOneProjectByIdAndRoleRow) (*database.Project, error)
	DeleteOne(project_id string) error
	FindById(user_id string, project_id string) (*database.FindOneProjectByIdRow, error)
	FindByIdAndRole(user_id string, project_id string, roles []string) (*database.FindOneProjectByIdAndRoleRow, error)
	ListMyProjects(user_id string) ([]database.FindProjectsForUserRow, error)
}

type service struct {
	ctx context.Context
	logger *slog.Logger
	repository ProjectRepository
	user_service user.Service
	org_service organization.Service
}

func NewService(ctx context.Context, logger *slog.Logger, repository ProjectRepository, user_service user.Service, org_service organization.Service) Service {
	return &service{
		ctx,
		logger,
		repository,
		user_service,
		org_service,
	}
}

func (s *service) Create(user_id string, org_id string, project_name  string) (*database.Project, error) {
	user, err := s.user_service.FindById(user_id, false)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	org_uuid := pgtype.UUID{}
	org_uuid.Scan(org_id)

	return s.repository.CreateProjectWithMember(
		database.CreateProjectParams{
			OrgID: org_uuid,
			Name: project_name,
		},
		database.CreateProjectMemberParams{
			UserID: user.UserID,
			Role: pgtype.Text{String: "owner", Valid: true},
		},
	)
}

func (s *service) UpdateOne(dto *database.FindOneProjectByIdAndRoleRow) (*database.Project, error) {
	updated_at := pgtype.Timestamp{
		Time: time.Now(),
		Valid: true,
	}

	project, err := s.repository.UpdateOne(database.UpdateOneProjectParams{
		ProjectID: dto.ProjectID,
		Name: dto.Name.String,
		UpdatedAt: updated_at,
	})

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *service) DeleteOne(project_id string) error {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)

	return s.repository.DeleteProjectWithMembers(project_uuid)
}

func (s *service) FindById(user_id string, project_id string) (*database.FindOneProjectByIdRow, error) {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	project_res, err := s.repository.FindOneById(database.FindOneProjectByIdParams{
		ProjectID: project_uuid,
		UserID: user_uuid,
	})

	if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, nil
			} else {
				return nil, errors.New("unable to find project")
			}
	}

	return project_res, nil
}

func (s *service) FindByIdAndRole(user_id string, project_id string, roles []string) (*database.FindOneProjectByIdAndRoleRow, error) {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	project_res, err := s.repository.FindOneByIdAndRole(database.FindOneProjectByIdAndRoleParams{
		ProjectID: project_uuid,
		UserID: user_uuid,
	})

	if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, nil
			} else {
				return nil, errors.New("unable to find user")
			}
	}

	return project_res, nil
}

func (s *service) ListMyProjects(user_id string) ([]database.FindProjectsForUserRow, error) {
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	projectus, err := s.repository.FindForUser(user_uuid)
	if err != nil {
		errmsg := err.Error()
		if strings.Contains(errmsg, "no rows") {
			return []database.FindProjectsForUserRow{}, nil
		}
		return nil, errors.New("unable to find project users, db query failed")
	}

	return projectus, nil
}
