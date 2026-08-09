package deployment

import (
	"mime/multipart"

	"github.com/salmanrf/capybara-cloud/internal/database"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
)

type StubAppDeploymentRepository struct {
	create_return *database.ApplicationDeployment
	create_error error
	create_n_calls int
	create_call_args []database.CreateApplicationDeploymentParams
	create_instance_return *database.DeploymentInstance
	create_instance_error error
	create_instance_n_calls int
	create_instance_call_args []database.CreateDeploymentInstanceParams
	find_current_return *database.ApplicationDeployment
	find_current_error error
	find_current_n_calls int
	find_current_call_args []string
	update_status_return *database.UpdateDeploymentStatusRow
	update_status_error error
	update_status_n_calls int
	update_status_call_args []database.UpdateDeploymentStatusParams
	update_instance_status_return *database.UpdateDeploymentInstanceStatusRow
	update_instance_status_error error
	update_instance_status_n_calls int
	update_instance_status_call_args []database.UpdateDeploymentInstanceStatusParams
	update_instance_return *database.DeploymentInstance
	update_instance_error error
	update_instance_n_calls int
	update_instance_call_args []database.UpdateDeploymentInstanceParams
}

func (s *StubAppDeploymentRepository) Clear() {
	s.create_return = nil
	s.create_error = nil
	s.create_n_calls = 0
	s.create_call_args = nil
	s.create_instance_return = nil
	s.create_instance_error = nil
	s.create_instance_n_calls = 0
	s.create_instance_call_args = nil
	s.find_current_return = nil
	s.find_current_error = nil
	s.find_current_n_calls = 0
	s.find_current_call_args = nil
	s.update_status_return = nil
	s.update_status_call_args = nil
	s.update_status_n_calls = 0
	s.update_status_error = nil
	s.update_instance_status_return = nil
	s.update_instance_status_call_args = nil
	s.update_instance_status_n_calls = 0
	s.update_instance_status_error = nil
	s.update_instance_return = nil
	s.update_instance_error = nil
	s.update_instance_n_calls = 0
	s.update_instance_call_args = nil
}

func (s *StubAppDeploymentRepository) Create(params database.CreateApplicationDeploymentParams) (*database.ApplicationDeployment, error) {
	s.create_n_calls += 1
	s.create_call_args = append(s.create_call_args, params)
	return s.create_return, s.create_error
}

func (s *StubAppDeploymentRepository) CreateInstance(params database.CreateDeploymentInstanceParams) (*database.DeploymentInstance, error) {
	s.create_instance_n_calls += 1
	s.create_instance_call_args = append(s.create_instance_call_args, params)
	return s.create_instance_return, s.create_instance_error
}

func (s *StubAppDeploymentRepository) FindCurrent(app_id string) (*database.ApplicationDeployment, error) {
	s.find_current_n_calls += 1
	s.find_current_call_args = append(s.find_current_call_args, app_id)
	return s.find_current_return, s.find_current_error
}

func (s *StubAppDeploymentRepository) UpdateStatus(params database.UpdateDeploymentStatusParams) (*database.UpdateDeploymentStatusRow, error) {
	s.update_status_n_calls += 1
	s.update_status_call_args = append(s.update_status_call_args, params)
	return s.update_status_return, s.update_status_error
}

func (s *StubAppDeploymentRepository) UpdateInstance(params database.UpdateDeploymentInstanceParams) (*database.DeploymentInstance, error) {
	s.update_instance_n_calls += 1
	s.update_instance_call_args = append(s.update_instance_call_args, params)
	return s.update_instance_return, s.update_instance_error
}

func (s *StubAppDeploymentRepository) UpdateInstanceStatus(params database.UpdateDeploymentInstanceStatusParams) (*database.UpdateDeploymentInstanceStatusRow, error) {
	s.update_instance_status_n_calls += 1
	s.update_instance_status_call_args = append(s.update_instance_status_call_args, params)
	return s.update_instance_status_return, s.update_instance_status_error
}

type StubService struct {
	deploy_return           *database.ApplicationDeployment
	deploy_error            error
	deploy_n_calls          int
	deploy_call_args        []deployCallArgs

	extract_return    shared_deployment.DeployStepResult
	extract_error     error
	extract_n_calls   int
	extract_call_args []shared_deployment.DeployRequest

	build_return    shared_deployment.DeployStepResult
	build_error     error
	build_n_calls   int
	build_call_args []shared_deployment.DeployRequest

	push_return    shared_deployment.DeployStepResult
	push_error     error
	push_n_calls   int
	push_call_args []shared_deployment.DeployRequest

	create_instance_return    *database.DeploymentInstance
	create_instance_error     error
	create_instance_n_calls   int
	create_instance_call_args []shared_deployment.DeployRequest

	update_instance_return    *database.DeploymentInstance
	update_instance_error     error
	update_instance_n_calls   int
	update_instance_call_args []*database.DeploymentInstance
}

type deployCallArgs struct {
	user_id               string
	app_id                string
	bundle_file_headers   *multipart.FileHeader
	bundle                multipart.File
}

func (s *StubService) Clear() {
	s.deploy_return = nil
	s.deploy_error = nil
	s.deploy_n_calls = 0
	s.deploy_call_args = nil

	s.extract_return = shared_deployment.DeployStepResult{}
	s.extract_error = nil
	s.extract_n_calls = 0
	s.extract_call_args = nil

	s.build_return = shared_deployment.DeployStepResult{}
	s.build_error = nil
	s.build_n_calls = 0
	s.build_call_args = nil

	s.push_return = shared_deployment.DeployStepResult{}
	s.push_error = nil
	s.push_n_calls = 0
	s.push_call_args = nil

	s.create_instance_return = nil
	s.create_instance_error = nil
	s.create_instance_n_calls = 0
	s.create_instance_call_args = nil

	s.update_instance_return = nil
	s.update_instance_error = nil
	s.update_instance_n_calls = 0
	s.update_instance_call_args = nil
}

func (s *StubService) Deploy(user_id string, app_id string, bundle_file_headers *multipart.FileHeader, bundle multipart.File) (*database.ApplicationDeployment, error) {
	s.deploy_n_calls += 1
	s.deploy_call_args = append(s.deploy_call_args, deployCallArgs{
		user_id:             user_id,
		app_id:              app_id,
		bundle_file_headers: bundle_file_headers,
		bundle:              bundle,
	})
	return s.deploy_return, s.deploy_error
}

func (s *StubService) createInstance(dto shared_deployment.DeployRequest) (*database.DeploymentInstance, error) {
	s.create_instance_n_calls += 1
	s.create_instance_call_args = append(s.create_instance_call_args, dto)

	return s.create_instance_return, s.create_instance_error
}

func (s *StubService) updateInstance(instance *database.DeploymentInstance) (*database.DeploymentInstance, error) {
	s.update_instance_n_calls += 1
	s.update_instance_call_args = append(s.update_instance_call_args, instance)

	return s.update_instance_return, s.update_instance_error
}

func (s *StubService) Extract(dto shared_deployment.DeployRequest) (shared_deployment.DeployStepResult, error) {
	s.extract_n_calls += 1
	s.extract_call_args = append(s.extract_call_args, dto)

	return s.extract_return, s.extract_error
}

func (s *StubService) Build(dto shared_deployment.DeployRequest) (shared_deployment.DeployStepResult, error) {
	s.build_n_calls += 1
	s.build_call_args = append(s.build_call_args, dto)

	return s.build_return, s.build_error
}

func (s *StubService) Push(dto shared_deployment.DeployRequest) (shared_deployment.DeployStepResult, error) {
	s.push_n_calls += 1
	s.push_call_args = append(s.push_call_args, dto)

	return s.push_return, s.push_error
}

type StubMasbroService struct {
	start_return    database.DeploymentInstance
	start_error     error
	start_n_calls   int
	start_call_args []shared_deployment.DeployRequest
}

func (s *StubMasbroService) Start(dto shared_deployment.DeployRequest, ins database.DeploymentInstance) (database.DeploymentInstance, error) {
	s.start_n_calls += 1
	s.start_call_args = append(s.start_call_args, dto)

	return s.start_return, s.start_error
}