package project

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/user"
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
	conn *pgxpool.Pool
	queries *database.Queries
	user_service user.Service
}

func NewService(ctx context.Context, conn *pgxpool.Pool, queries *database.Queries, user_service user.Service) Service {
	return &service{
		ctx,
		conn,
		queries,
		user_service,
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

	trx, err := s.conn.Begin(s.ctx)
	if err != nil {
		return nil, err
	}
	defer trx.Rollback(s.ctx)
	q := s.queries.WithTx(trx)

	org_uuid := pgtype.UUID{}
	org_uuid.Scan(org_id)
	
	project, err := q.CreateProject(s.ctx, database.CreateProjectParams{
		OrgID: org_uuid,
		Name: project_name,
	})

	if err != nil {
		fmt.Println("Error creating project", user.UserID, project_name, err)
		return nil, err
	}

	_, err = q.CreateProjectMember(
		s.ctx, 
		database.CreateProjectMemberParams{
			ProjectID: project.ProjectID,
			UserID: user.UserID,
			Role: "owner",
		},
	)

	if err != nil {
		fmt.Println("Error creating project", user.UserID, project_name, err)
		return nil, err
	}

	trx.Commit(s.ctx)
	
	return &project, nil
}

func (s *service) UpdateOne(dto *database.FindOneProjectByIdAndRoleRow) (*database.Project, error) {
	updated_at := pgtype.Timestamp{
		Time: time.Now(),
		Valid: true,
	}
	
	project, err := s.queries.UpdateOneProject(s.ctx, database.UpdateOneProjectParams{
		ProjectID: dto.ProjectID,
		Name: dto.Name.String,
		UpdatedAt: updated_at,
	})

	if err != nil {
		fmt.Println("Error updating project", project.ProjectID.String(), project.Name, err)
		return nil, err
	}

	return &project, nil
}

func (s *service) DeleteOne(project_id string) error {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)

	trx, err := s.conn.Begin(s.ctx)
	defer trx.Rollback(s.ctx)
	q := s.queries.WithTx(trx)

	if err != nil {
		return err
	}

	err = s.DeleteProjectMembers(trx, project_id)

	if err != nil {
		return err
	}

	err = q.DeleteOneProject(s.ctx, project_uuid)

	if err != nil {
		return err
	}

	err = trx.Commit(s.ctx)
	
	return err
}

func (s *service) DeleteProjectMembers(trx pgx.Tx, project_id string) error {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)

	q := s.queries.WithTx(trx)

	err := q.DeleteProjectMembersByProjectId(s.ctx, project_uuid)

	return err
}

func (s *service) FindById(user_id string, project_id string) (*database.FindOneProjectByIdRow, error) {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	project_res, err := s.queries.FindOneProjectById(s.ctx, database.FindOneProjectByIdParams{
		ProjectID: project_uuid,
		UserID: user_uuid,
	})

	if err != nil {
		err_msg := err.Error()
			if strings.Contains(err_msg, "no rows") {
				return nil, nil
			} else {
				fmt.Println("Error finding project", err_msg)
				return nil, errors.New("unable to find project")
			}
	}

	return &project_res, nil
}

func (s *service) FindByIdAndRole(user_id string, project_id string, roles []string) (*database.FindOneProjectByIdAndRoleRow, error) {
	project_uuid := pgtype.UUID{}
	project_uuid.Scan(project_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	project_res, err := s.queries.FindOneProjectByIdAndRole(s.ctx, database.FindOneProjectByIdAndRoleParams{
		ProjectID: project_uuid,
		UserID: user_uuid,
	})

	if err != nil {
		err_msg := err.Error()
			if strings.Contains(err_msg, "no rows") {
				return nil, nil
			} else {
				fmt.Println("Error finding user", err_msg)
				return nil, errors.New("unable to find user")
			}
	}

	return &project_res, nil
}

func (s *service) ListMyProjects(user_id string) ([]database.FindProjectsForUserRow, error) {
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	projectus, err := s.queries.FindProjectsForUser(s.ctx, user_uuid)
	if err != nil {
		errmsg := err.Error()		
		fmt.Println("Error at project_service.ListMyProjects", errmsg)
		if strings.Contains(errmsg, "no rows") {
			return []database.FindProjectsForUserRow{}, nil
		}
		return nil, errors.New("unable to find project users, db query failed")
	}

	return projectus, nil
}