package application

import (
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
)

type StubProjectService struct {
	create_return *database.Project
	create_err    error
	update_one_return *database.Project
	update_one_err    error
	find_by_id_return *database.FindOneProjectByIdRow
	find_by_id_error  error
	find_by_id_and_role_return *database.FindOneProjectByIdAndRoleRow
	find_by_id_and_role_error  error
	delete_one_err error
}

func (s *StubProjectService) Create(user_id string, org_id, project_name string) (*database.Project, error) {
	return s.create_return, s.create_err
}

func (s *StubProjectService) UpdateOne(dto *database.FindOneProjectByIdAndRoleRow) (*database.Project, error) {
	return s.update_one_return, s.update_one_err
}

func (s *StubProjectService) DeleteOne(project_id string) error {
	return s.delete_one_err
}

func (s *StubProjectService) FindById(user_id string, project_id string) (*database.FindOneProjectByIdRow, error) {
	return s.find_by_id_return, s.find_by_id_error
}

func (s *StubProjectService) FindByIdAndRole(user_id string, project_id string, roles []string) (*database.FindOneProjectByIdAndRoleRow, error) {
	return s.find_by_id_and_role_return, s.find_by_id_and_role_error
}

func (s *StubProjectService) ListMyProjects(user_id string) ([]database.FindProjectsForUserRow, error) {
	return []database.FindProjectsForUserRow{}, nil
}

type StubApplicationRepository struct {
	find_one_with_project_member_return *database.FindOneApplicationWithProjectMemberRow
	find_one_with_project_member_error error
	upsert_config_return *database.ApplicationConfig
	upsert_config_error error
	create_application_return *database.Application
	create_application_error error
	update_one_application_return *database.Application
	update_one_application_error error
	find_one_with_project_member_n_calls int
	find_one_with_project_member_call_args []database.FindOneApplicationWithProjectMemberParams
	upsert_config_n_calls int
	upsert_config_call_args []database.CreateApplicationConfigParams
	create_application_n_calls int
	create_application_call_args []database.CreateApplicationParams
	update_one_application_n_calls int
	update_one_application_call_args []database.UpdateOneApplicationParams
}

func (s *StubApplicationRepository) Clear() {
	s.find_one_with_project_member_return = nil
	s.find_one_with_project_member_error = nil
	s.upsert_config_return = nil
	s.upsert_config_error = nil
	s.create_application_return = nil
	s.create_application_error = nil
	s.update_one_application_return = nil
	s.update_one_application_error = nil
	s.find_one_with_project_member_n_calls = 0
	s.find_one_with_project_member_call_args = nil
	s.upsert_config_n_calls = 0
	s.upsert_config_call_args = nil
	s.create_application_n_calls = 0
	s.create_application_call_args = nil
	s.update_one_application_n_calls = 0
	s.update_one_application_call_args = nil
}

func (s *StubApplicationRepository) FindOneWithProjectMember(params database.FindOneApplicationWithProjectMemberParams) (*database.FindOneApplicationWithProjectMemberRow, error) {
	s.find_one_with_project_member_n_calls += 1
	s.find_one_with_project_member_call_args = append(s.find_one_with_project_member_call_args, params)
	return s.find_one_with_project_member_return, s.find_one_with_project_member_error
}

func (s *StubApplicationRepository) UpsertConfig(params database.CreateApplicationConfigParams) (*database.ApplicationConfig, error) {
	s.upsert_config_n_calls += 1
	s.upsert_config_call_args = append(s.upsert_config_call_args, params)
	return s.upsert_config_return, s.upsert_config_error
}

func (s *StubApplicationRepository) CreateApplication(params database.CreateApplicationParams) (*database.Application, error) {
	s.create_application_n_calls += 1
	s.create_application_call_args = append(s.create_application_call_args, params)
	return s.create_application_return, s.create_application_error
}

func (s *StubApplicationRepository) UpdateOneApplication(params database.UpdateOneApplicationParams) (*database.Application, error) {
	s.update_one_application_n_calls += 1
	s.update_one_application_call_args = append(s.update_one_application_call_args, params)
	return s.update_one_application_return, s.update_one_application_error
}

type StubApplicationService struct {
	Find_one_return *database.FindOneApplicationWithProjectMemberRow
	Find_one_error  error
	Find_one_n_calls int
	Find_one_call_args []FindOneCallArgs

	create_return *database.Application
	create_error  error
	create_n_calls int

	update_return *database.Application
	update_error  error
	update_n_calls int

	create_config_return *database.ApplicationConfig
	create_config_error  error
	create_config_n_calls int

	find_one_config_return *dto.ApplicationConfigResponse
	find_one_config_error  error
	find_one_config_n_calls int
}

type FindOneCallArgs struct {
	app_id  string
	user_id string
}

func (s *StubApplicationService) Clear() {
	s.Find_one_return = nil
	s.Find_one_error = nil
	s.Find_one_n_calls = 0
	s.Find_one_call_args = nil

	s.create_return = nil
	s.create_error = nil
	s.create_n_calls = 0

	s.update_return = nil
	s.update_error = nil
	s.update_n_calls = 0

	s.create_config_return = nil
	s.create_config_error = nil
	s.create_config_n_calls = 0

	s.find_one_config_return = nil
	s.find_one_config_error = nil
	s.find_one_config_n_calls = 0
}

func (s *StubApplicationService) Create(user_id string, create_dto dto.CreateApplicationDto) (*database.Application, error) {
	s.create_n_calls += 1
	return s.create_return, s.create_error
}

func (s *StubApplicationService) Update(app_id string, user_id string, update_dto dto.UpdateApplicationDto) (*database.Application, error) {
	s.update_n_calls += 1
	return s.update_return, s.update_error
}

func (s *StubApplicationService) FindOne(app_id string, user_id string) (*database.FindOneApplicationWithProjectMemberRow, error) {
	s.Find_one_n_calls += 1
	s.Find_one_call_args = append(s.Find_one_call_args, FindOneCallArgs{app_id, user_id})
	return s.Find_one_return, s.Find_one_error
}

func (s *StubApplicationService) CreateConfig(app_id string, user_id string, create_dto dto.CreateApplicationConfigDto) (*database.ApplicationConfig, error) {
	s.create_config_n_calls += 1
	return s.create_config_return, s.create_config_error
}

func (s *StubApplicationService) FindOneConfig(app_id string, user_id string) (*dto.ApplicationConfigResponse, error) {
	s.find_one_config_n_calls += 1
	return s.find_one_config_return, s.find_one_config_error
}