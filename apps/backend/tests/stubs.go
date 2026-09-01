package tests

import (
	"mime/multipart"
	"net/http"
	"time"

	"github.com/salmanrf/capybara-cloud/apps/backend/internal/deployment"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type api_server struct {
		http.Handler
}

type StubUserService struct {
	find_by_id_n_calls int
	create_n_calls int
	find_by_id_return *database.User
	find_by_id_err error
	create_return *database.User
	create_err error
}

type StubAuthService struct {} 

type StubOrgService struct {
	create_return *database.Organization
	create_err error
	update_one_return *database.Organization
	update_one_err error
	find_by_id_return *database.FindOneOrganizationByIdRow
	find_by_id_error error
	find_by_id_and_role_return *database.FindOneOrganizationByIdAndRoleRow
	find_by_id_and_role_error error
	update_one_n_calls int
	update_one_call_args []*database.FindOneOrganizationByIdAndRoleRow 
	delete_one_n_calls int
	delete_one_call_args []string
	delete_one_err error
}

func (s *StubUserService) FindById(identifier string, is_email bool) (*database.User, error) {
	s.find_by_id_n_calls += 1
	return s.find_by_id_return, s.find_by_id_err
}

func (s *StubUserService) Create(create_dto dto.SignupDto) (*database.User, error) {
	s.create_n_calls += 1
	return s.create_return, s.create_err
} 

func (s *StubAuthService) GetMe(user_id string) (*database.User, error) {
	return nil, nil
}

func (s *StubOrgService) Create(user_id string, org_name string) (*database.Organization, error) {
	return s.create_return, s.create_err
}

func (s *StubOrgService) UpdateOne(dto *database.FindOneOrganizationByIdAndRoleRow) (*database.Organization, error) {
	s.update_one_n_calls += 1
	s.update_one_call_args = append(s.update_one_call_args, dto) 
	
	return s.update_one_return, s.update_one_err
}

func (s *StubOrgService) DeleteOne(org_id string) error {
	s.delete_one_n_calls += 1
	s.delete_one_call_args = append(s.delete_one_call_args, org_id)
	
	return s.delete_one_err
}

func (s *StubOrgService) FindById(user_id string, org_id string) (*database.FindOneOrganizationByIdRow, error) {
	return s.find_by_id_return, s.find_by_id_error
}

func (s *StubOrgService) FindByIdAndRole(user_id string, org_id string, roles []string) (*database.FindOneOrganizationByIdAndRoleRow, error) {
	return s.find_by_id_and_role_return, s.find_by_id_and_role_error
}

func (s *StubOrgService) ListMyOrgs(user_id string) ([]database.FindOrganizationsForUserRow, error) {
	return []database.FindOrganizationsForUserRow{}, nil
} 

type StubProjectService struct {
	create_n_calls int
	create_call_args [][]string
	create_return *database.Project
	create_err error
	update_one_return *database.Project
	update_one_err error
	find_by_id_return *database.FindOneProjectByIdRow
	find_by_id_error error
	find_by_id_and_role_return *database.FindOneProjectByIdAndRoleRow
	find_by_id_and_role_error error
	update_one_n_calls int
	update_one_call_args []*database.FindOneProjectByIdAndRoleRow 
	delete_one_n_calls int
	delete_one_call_args []string
	delete_one_err error
} 

func (s *StubProjectService) Create(user_id string, org_id, project_name string) (*database.Project, error) {
	s.create_n_calls += 1
	s.create_call_args = append(s.create_call_args, []string{user_id, org_id, project_name}) 
	
	return s.create_return, s.create_err
}

func (s *StubProjectService) UpdateOne(dto *database.FindOneProjectByIdAndRoleRow) (*database.Project, error) {
	s.update_one_n_calls += 1
	s.update_one_call_args = append(s.update_one_call_args, dto) 
	
	return s.update_one_return, s.update_one_err
}

func (s *StubProjectService) DeleteOne(org_id string) error {
	s.delete_one_n_calls += 1
	s.delete_one_call_args = append(s.delete_one_call_args, org_id)
	
	return s.delete_one_err
}

func (s *StubProjectService) FindById(user_id string, org_id string) (*database.FindOneProjectByIdRow, error) {
	return s.find_by_id_return, s.find_by_id_error
}

func (s *StubProjectService) FindByIdAndRole(user_id string, org_id string, roles []string) (*database.FindOneProjectByIdAndRoleRow, error) {
	return s.find_by_id_and_role_return, s.find_by_id_and_role_error
}

func (s *StubProjectService) ListMyProjects(user_id string) ([]database.FindProjectsForUserRow, error) {
	return []database.FindProjectsForUserRow{}, nil
}

type StubApplicationService struct {
	Create_n_calls int
	Create_return *database.Application
	Create_err error
	Create_config_calls_arg1 []string
	Create_config_calls_arg2 []string
	Create_config_calls_arg3 []dto.CreateApplicationConfigDto
	Create_config_n_calls int
	Create_config_return *database.ApplicationConfig
	Create_config_err error
	Create_calls_arg1 []string
	Update_n_calls int
	Update_return *database.Application
	Update_err error
	Update_calls_arg1 []string
	Update_calls_arg2 []string
	Update_calls_arg3 []dto.UpdateApplicationDto
	Find_one_n_calls int
	Find_one_return *database.FindOneApplicationWithProjectMemberRow
	Find_one_error error
	Find_one_calls_arg1 []string
	Find_one_calls_arg2 []string
	Find_one_config_calls_arg1 []string
	Find_one_config_calls_arg2 []string
	Find_one_config_n_calls int
	Find_one_config_return *dto.ApplicationConfigResponse
	Find_one_config_error error
}

func (s *StubApplicationService) Clear() {
	s.Create_n_calls = 0
	s.Create_return = nil
	s.Create_err = nil
	s.Update_n_calls = 0
	s.Update_return = nil
	s.Update_err = nil
	s.Create_calls_arg1 = []string{}
	s.Update_calls_arg1 = []string{}
	s.Update_calls_arg2 = []string{}
	s.Update_calls_arg3 = []dto.UpdateApplicationDto{}
	s.Find_one_return = nil
	s.Find_one_error = nil
	s.Find_one_n_calls = 0
	s.Find_one_calls_arg1 = []string{}
	s.Find_one_calls_arg2 = []string{}
	s.Create_config_n_calls = 0
	s.Create_config_calls_arg1 = []string{}
	s.Create_config_calls_arg2 = []string{}
	s.Create_config_calls_arg3 = []dto.CreateApplicationConfigDto{}
	s.Create_config_return = nil
	s.Create_config_err = nil
	s.Find_one_config_n_calls = 0
	s.Find_one_config_calls_arg1 = []string{}
	s.Find_one_config_calls_arg2 = []string{}
	s.Find_one_config_return = nil
	s.Find_one_config_error = nil
}

func (s *StubApplicationService) Create(user_id string, dto dto.CreateApplicationDto) (*database.Application, error) {
	s.Create_n_calls += 1
	s.Create_calls_arg1 = append(s.Create_calls_arg1, user_id)
	return s.Create_return, s.Create_err
}

func (s *StubApplicationService) Update(app_id string, user_id string, dto dto.UpdateApplicationDto) (*database.Application, error) {
	s.Update_n_calls += 1
	s.Update_calls_arg1 = append(s.Update_calls_arg1, app_id)
	s.Update_calls_arg2 = append(s.Update_calls_arg2, user_id)
	s.Update_calls_arg3 = append(s.Update_calls_arg3, dto)
	return s.Update_return, s.Update_err
}

func (s *StubApplicationService) FindOne(app_id string, user_id string) (*database.FindOneApplicationWithProjectMemberRow, error) {
	s.Find_one_n_calls += 1
	s.Find_one_calls_arg1 = append(s.Find_one_calls_arg1, app_id)
	s.Find_one_calls_arg2 = append(s.Find_one_calls_arg2, user_id)
	return s.Find_one_return, s.Find_one_error
} 

func (s *StubApplicationService) CreateConfig(app_id string, user_id string, dto dto.CreateApplicationConfigDto) (*database.ApplicationConfig, error) {
	s.Create_config_n_calls += 1
	s.Create_config_calls_arg1 = append(s.Create_config_calls_arg1, app_id)
	s.Create_config_calls_arg2 = append(s.Create_config_calls_arg2, user_id)
	s.Create_config_calls_arg3 = append(s.Create_config_calls_arg3, dto)
	return s.Create_config_return, s.Create_config_err
}

func (s *StubApplicationService) FindOneConfig(app_id string, user_id string) (*dto.ApplicationConfigResponse, error) {
	s.Find_one_config_n_calls += 1
	s.Find_one_config_calls_arg1 = append(s.Find_one_config_calls_arg1, app_id)
	s.Find_one_config_calls_arg2 = append(s.Find_one_config_calls_arg2, user_id)
	return s.Find_one_config_return, s.Find_one_config_error
}

type StubDeploymentService struct {
	deployment.Service

	deploy_n_calls int
	deploy_return *database.ApplicationDeployment
	deploy_err error
	deploy_calls_arg1 []string
	deploy_calls_arg2 []string
	deploy_calls_arg3 []*multipart.FileHeader
	deploy_calls_arg4 []multipart.File

	extract_n_calls   int
	extract_return    deployment.DeployStepResult
	extract_err       error
	extract_call_args []deployment.DeployRequest

	build_n_calls   int
	build_return    deployment.DeployStepResult
	build_err       error
	build_call_args []deployment.DeployRequest

	push_n_calls   int
	push_return    deployment.DeployStepResult
	push_err       error
	push_call_args []deployment.DeployRequest
}

func (s *StubDeploymentService) Clear() {
	s.deploy_n_calls = 0
	s.deploy_return = nil
	s.deploy_err = nil
	s.deploy_calls_arg1 = []string{}
	s.deploy_calls_arg2 = []string{}
	s.deploy_calls_arg3 = []*multipart.FileHeader{}
	s.deploy_calls_arg4 = []multipart.File{}

	s.extract_n_calls = 0
	s.extract_return = deployment.DeployStepResult{}
	s.extract_err = nil
	s.extract_call_args = []deployment.DeployRequest{}

	s.build_n_calls = 0
	s.build_return = deployment.DeployStepResult{}
	s.build_err = nil
	s.build_call_args = []deployment.DeployRequest{}

	s.push_n_calls = 0
	s.push_return = deployment.DeployStepResult{}
	s.push_err = nil
	s.push_call_args = []deployment.DeployRequest{}
}

func (s *StubDeploymentService) Deploy(user_id string, application_id string, bundle_file_headers *multipart.FileHeader, bundle multipart.File) (*database.ApplicationDeployment, error) {
	s.deploy_n_calls += 1
	s.deploy_calls_arg1 = append(s.deploy_calls_arg1, user_id)
	s.deploy_calls_arg2 = append(s.deploy_calls_arg2, application_id)
	s.deploy_calls_arg3 = append(s.deploy_calls_arg3, bundle_file_headers)
	s.deploy_calls_arg4 = append(s.deploy_calls_arg4, bundle)
	return s.deploy_return, s.deploy_err
}

func (s *StubDeploymentService) Extract(dto deployment.DeployRequest) (deployment.DeployStepResult, error) {
	s.extract_n_calls += 1
	s.extract_call_args = append(s.extract_call_args, dto)

	result := s.extract_return
	result.DeploymentDto = dto.DeploymentDto

	return result, s.extract_err
}

func (s *StubDeploymentService) Build(dto deployment.DeployRequest) (deployment.DeployStepResult, error) {
	s.build_n_calls += 1
	s.build_call_args = append(s.build_call_args, dto)

	result := s.build_return
	result.DeploymentDto = dto.DeploymentDto

	return result, s.build_err
}

func (s *StubDeploymentService) Push(dto deployment.DeployRequest) (deployment.DeployStepResult, error) {
	s.push_n_calls += 1
	s.push_call_args = append(s.push_call_args, dto)

	result := s.push_return
	result.DeploymentDto = dto.DeploymentDto

	return result, s.push_err
}

type StubJwtValidator struct {
	validate_return string
	validate_error error
	make_return string
	make_error error
}

func (v *StubJwtValidator) ValidateJWT(token, secret string) (string, error) {
	return v.validate_return, v.validate_error
}

func (v *StubJwtValidator) MakeJWT(sub string, jwt_secret string, expires_in time.Duration) (string, error) {
	return v.make_return, v.make_error
}



