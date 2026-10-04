package project

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/organization"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/user"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type Service interface {
	Create(user_id string, org_id string, project_name string) (*database.Project, error)
	UpdateOne(user_id string, project_id string, project_name string) (*database.Project, error)
	DeleteOne(user_id string, project_id string) error
	FindById(user_id string, project_id string) (*database.FindOneProjectByIdRow, error)
	FindByIdAndRole(user_id string, project_id string, roles []string) (*database.FindOneProjectByIdAndRoleRow, error)
	ListMyProjects(user_id string, org_id string) ([]database.FindProjectsForUserRow, error)
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

	org, err := s.org_service.FindByIdAndRole(user_id, org_id, []string{})
	if err != nil || org == nil {
		return nil, errors.New("invalid_role")
	}
	allowed_roles := []string{
		"owner",
		"editor",
	}
	if !slices.Contains(allowed_roles, org.Role) {
		return nil, errors.New("invalid_role")
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

func (s *service) UpdateOne(user_id string, project_id string, project_name string) (*database.Project, error) {
	existing, err := s.FindByIdAndRole(user_id, project_id, []string{})
	if err != nil || existing == nil {
		return nil, errors.New("invalid_role")
	}

	allowed_roles := []string{
		"owner",
		"editor",
	}
	if !slices.Contains(allowed_roles, existing.Role.String) {
		return nil, errors.New("invalid_role")
	}

	updated_at := pgtype.Timestamp{
		Time: time.Now(),
		Valid: true,
	}

	project, err := s.repository.UpdateOne(database.UpdateOneProjectParams{
		ProjectID: existing.ProjectID,
		Name: project_name,
		UpdatedAt: updated_at,
	})

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *service) DeleteOne(user_id string, project_id string) error {
	existing, err := s.FindByIdAndRole(user_id, project_id, []string{})
	if err != nil || existing == nil {
		return errors.New("invalid_role")
	}

	allowed_roles := []string{
		"owner",
		"editor",
	}
	if !slices.Contains(allowed_roles, existing.Role.String) {
		return errors.New("invalid_role")
	}

	return s.repository.DeleteProjectWithMembers(existing.ProjectID)
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

func (s *service) ListMyProjects(user_id string, org_id string) ([]database.FindProjectsForUserRow, error) {
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	org_uuid := pgtype.UUID{}
	if org_id != "" {
		org_uuid.Scan(org_id)
	}

	projectus, err := s.repository.FindForUser(database.FindProjectsForUserParams{
		UserID: user_uuid,
		OrgID:  org_uuid,
	})
	if err != nil {
		errmsg := err.Error()
		if strings.Contains(errmsg, "no rows") {
			return []database.FindProjectsForUserRow{}, nil
		}
		return nil, errors.New("unable to find project users, db query failed")
	}

	return projectus, nil
}
