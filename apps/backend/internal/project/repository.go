package project

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

// @params none
// @return none
// repository is the sqlc-backed ProjectRepository; conn is used only to open transactions for the atomic methods.
type repository struct {
	ctx     context.Context
	conn    *pgxpool.Pool
	queries *database.Queries
}

type ProjectRepository interface {
	CreateProjectWithMember(project database.CreateProjectParams, member database.CreateProjectMemberParams) (*database.Project, error)
	DeleteProjectWithMembers(project_id pgtype.UUID) error
	FindOneById(params database.FindOneProjectByIdParams) (*database.FindOneProjectByIdRow, error)
	FindOneByIdAndRole(params database.FindOneProjectByIdAndRoleParams) (*database.FindOneProjectByIdAndRoleRow, error)
	FindForUser(user_id pgtype.UUID) ([]database.FindProjectsForUserRow, error)
	UpdateOne(params database.UpdateOneProjectParams) (*database.Project, error)
}

func NewRepository(ctx context.Context, conn *pgxpool.Pool, queries *database.Queries) ProjectRepository {
	return &repository{
		ctx:     ctx,
		conn:    conn,
		queries: queries,
	}
}

func (r *repository) CreateProjectWithMember(project database.CreateProjectParams, member database.CreateProjectMemberParams) (*database.Project, error) {
	trx, err := r.conn.Begin(r.ctx)
	if err != nil {
		return nil, err
	}
	defer trx.Rollback(r.ctx)

	q := r.queries.WithTx(trx)

	created, err := q.CreateProject(r.ctx, project)
	if err != nil {
		return nil, err
	}

	member.ProjectID = created.ProjectID

	_, err = q.CreateProjectMember(r.ctx, member)
	if err != nil {
		return nil, err
	}

	err = trx.Commit(r.ctx)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *repository) DeleteProjectWithMembers(project_id pgtype.UUID) error {
	trx, err := r.conn.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer trx.Rollback(r.ctx)

	q := r.queries.WithTx(trx)

	err = q.DeleteProjectMembersByProjectId(r.ctx, project_id)
	if err != nil {
		return err
	}

	err = q.DeleteOneProject(r.ctx, project_id)
	if err != nil {
		return err
	}

	return trx.Commit(r.ctx)
}

func (r *repository) FindOneById(params database.FindOneProjectByIdParams) (*database.FindOneProjectByIdRow, error) {
	project, err := r.queries.FindOneProjectById(r.ctx, params)

	return &project, err
}

func (r *repository) FindOneByIdAndRole(params database.FindOneProjectByIdAndRoleParams) (*database.FindOneProjectByIdAndRoleRow, error) {
	project, err := r.queries.FindOneProjectByIdAndRole(r.ctx, params)

	return &project, err
}

func (r *repository) FindForUser(user_id pgtype.UUID) ([]database.FindProjectsForUserRow, error) {
	return r.queries.FindProjectsForUser(r.ctx, user_id)
}

func (r *repository) UpdateOne(params database.UpdateOneProjectParams) (*database.Project, error) {
	project, err := r.queries.UpdateOneProject(r.ctx, params)

	return &project, err
}
