package tests

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
)

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

type StubJwtValidator struct {
	validate_return string
	validate_error error 
	make_return string
	make_error error 
}

func (v *StubJwtValidator) ValidateJWT(token, secret string) (string, error) {
	return v.validate_return, v.validate_error
}

func (v *StubJwtValidator) MakeJWT(user_id pgtype.UUID, jwt_secret string, expires_in time.Duration) (string, error) {
	return v.make_return, v.make_error
}



