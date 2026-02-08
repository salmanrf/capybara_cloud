package application

import (
	"github.com/salmanrf/capybara-cloud/internal/database"
)

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