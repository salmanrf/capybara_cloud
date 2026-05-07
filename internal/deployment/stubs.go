package deployment

import "github.com/salmanrf/capybara-cloud/internal/database"

type StubAppDeploymentRepository struct {
	create_return *database.ApplicationDeployment
	create_error error
	create_n_calls int
	create_call_args []database.CreateApplicationDeploymentParams
}

func (s *StubAppDeploymentRepository) Clear() {
	s.create_return = nil
	s.create_error = nil
	s.create_n_calls = 0
	s.create_call_args = nil
}

func (s *StubAppDeploymentRepository) Create(params database.CreateApplicationDeploymentParams) (*database.ApplicationDeployment, error) {
	s.create_n_calls += 1
	s.create_call_args = append(s.create_call_args, params)
	return s.create_return, s.create_error
}