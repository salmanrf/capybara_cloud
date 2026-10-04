package project

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type CreateProjectWithMemberCallArgs struct {
	project database.CreateProjectParams
	member  database.CreateProjectMemberParams
}

type StubProjectRepository struct {
	create_project_with_member_return    *database.Project
	create_project_with_member_err       error
	create_project_with_member_n_calls   int
	create_project_with_member_call_args []CreateProjectWithMemberCallArgs

	delete_project_with_members_err       error
	delete_project_with_members_n_calls   int
	delete_project_with_members_call_args []pgtype.UUID

	find_one_by_id_return    *database.FindOneProjectByIdRow
	find_one_by_id_err       error
	find_one_by_id_n_calls   int
	find_one_by_id_call_args []database.FindOneProjectByIdParams

	find_one_by_id_and_role_return    *database.FindOneProjectByIdAndRoleRow
	find_one_by_id_and_role_err       error
	find_one_by_id_and_role_n_calls   int
	find_one_by_id_and_role_call_args []database.FindOneProjectByIdAndRoleParams

	find_for_user_return    []database.FindProjectsForUserRow
	find_for_user_err       error
	find_for_user_n_calls   int
	find_for_user_call_args []database.FindProjectsForUserParams

	update_one_return    *database.Project
	update_one_err       error
	update_one_n_calls   int
	update_one_call_args []database.UpdateOneProjectParams
}

func (s *StubProjectRepository) Clear() {
	*s = StubProjectRepository{}
}

func (s *StubProjectRepository) CreateProjectWithMember(project database.CreateProjectParams, member database.CreateProjectMemberParams) (*database.Project, error) {
	s.create_project_with_member_n_calls += 1
	s.create_project_with_member_call_args = append(s.create_project_with_member_call_args, CreateProjectWithMemberCallArgs{project, member})
	return s.create_project_with_member_return, s.create_project_with_member_err
}

func (s *StubProjectRepository) DeleteProjectWithMembers(project_id pgtype.UUID) error {
	s.delete_project_with_members_n_calls += 1
	s.delete_project_with_members_call_args = append(s.delete_project_with_members_call_args, project_id)
	return s.delete_project_with_members_err
}

func (s *StubProjectRepository) FindOneById(params database.FindOneProjectByIdParams) (*database.FindOneProjectByIdRow, error) {
	s.find_one_by_id_n_calls += 1
	s.find_one_by_id_call_args = append(s.find_one_by_id_call_args, params)
	return s.find_one_by_id_return, s.find_one_by_id_err
}

func (s *StubProjectRepository) FindOneByIdAndRole(params database.FindOneProjectByIdAndRoleParams) (*database.FindOneProjectByIdAndRoleRow, error) {
	s.find_one_by_id_and_role_n_calls += 1
	s.find_one_by_id_and_role_call_args = append(s.find_one_by_id_and_role_call_args, params)
	return s.find_one_by_id_and_role_return, s.find_one_by_id_and_role_err
}

func (s *StubProjectRepository) FindForUser(params database.FindProjectsForUserParams) ([]database.FindProjectsForUserRow, error) {
	s.find_for_user_n_calls += 1
	s.find_for_user_call_args = append(s.find_for_user_call_args, params)
	return s.find_for_user_return, s.find_for_user_err
}

func (s *StubProjectRepository) UpdateOne(params database.UpdateOneProjectParams) (*database.Project, error) {
	s.update_one_n_calls += 1
	s.update_one_call_args = append(s.update_one_call_args, params)
	return s.update_one_return, s.update_one_err
}

type StubUserService struct {
	find_by_id_return *database.User
	find_by_id_err    error
}

func (s *StubUserService) Clear() {
	*s = StubUserService{}
}

func (s *StubUserService) FindById(identifier string, is_email bool) (*database.User, error) {
	return s.find_by_id_return, s.find_by_id_err
}

func (s *StubUserService) Create(create_params dto.SignupDto) (*database.User, error) {
	return nil, nil
}

type StubOrgService struct {
	create_return *database.Organization
	create_err    error

	update_one_return *database.Organization
	update_one_err    error

	delete_one_err error

	find_by_id_return  *database.FindOneOrganizationByIdRow
	find_by_id_err     error
	find_by_id_n_calls int

	find_by_id_and_role_return  *database.FindOneOrganizationByIdAndRoleRow
	find_by_id_and_role_err     error
	find_by_id_and_role_n_calls int

	list_my_orgs_return []database.FindOrganizationsForUserRow
	list_my_orgs_err    error
}

func (s *StubOrgService) Clear() {
	*s = StubOrgService{}
}

func (s *StubOrgService) Create(user_id string, org_name string) (*database.Organization, error) {
	return s.create_return, s.create_err
}

func (s *StubOrgService) UpdateOne(dto *database.FindOneOrganizationByIdAndRoleRow) (*database.Organization, error) {
	return s.update_one_return, s.update_one_err
}

func (s *StubOrgService) DeleteOne(org_id string) error {
	return s.delete_one_err
}

func (s *StubOrgService) FindById(user_id string, org_id string) (*database.FindOneOrganizationByIdRow, error) {
	s.find_by_id_n_calls += 1
	return s.find_by_id_return, s.find_by_id_err
}

func (s *StubOrgService) FindByIdAndRole(user_id string, org_id string, roles []string) (*database.FindOneOrganizationByIdAndRoleRow, error) {
	s.find_by_id_and_role_n_calls += 1
	return s.find_by_id_and_role_return, s.find_by_id_and_role_err
}

func (s *StubOrgService) ListMyOrgs(user_id string) ([]database.FindOrganizationsForUserRow, error) {
	return s.list_my_orgs_return, s.list_my_orgs_err
}
