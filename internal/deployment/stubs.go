package deployment

import "github.com/salmanrf/capybara-cloud/internal/database"

type StubAppDeploymentRepository struct {
	create_return *database.ApplicationDeployment
	create_error error
	create_n_calls int
	create_call_args []database.CreateApplicationDeploymentParams
	find_current_return *database.ApplicationDeployment
	find_current_error error
	find_current_n_calls int
	find_current_call_args []string
}

func (s *StubAppDeploymentRepository) Clear() {
	s.create_return = nil
	s.create_error = nil
	s.create_n_calls = 0
	s.create_call_args = nil
	s.find_current_return = nil
	s.find_current_error = nil
	s.find_current_n_calls = 0
	s.find_current_call_args = nil
}

func (s *StubAppDeploymentRepository) Create(params database.CreateApplicationDeploymentParams) (*database.ApplicationDeployment, error) {
	s.create_n_calls += 1
	s.create_call_args = append(s.create_call_args, params)
	return s.create_return, s.create_error
}

func (s *StubAppDeploymentRepository) FindCurrent(app_id string) (*database.ApplicationDeployment, error) {
	s.find_current_n_calls += 1
	s.find_current_call_args = append(s.find_current_call_args, app_id)
	return s.find_current_return, s.find_current_error
}