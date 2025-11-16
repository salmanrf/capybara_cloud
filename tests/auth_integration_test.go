package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salmanrf/capybara-cloud/api"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
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
type StubOrgService struct {} 

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
	return nil, nil
}

func (s *StubOrgService) ListMyOrgs(user_id string) ([]database.FindOrganizationsForUserRow, error) {
	return []database.FindOrganizationsForUserRow{}, nil
} 

func TestAuthSignupIntegration(t *testing.T) {
	test_ctx := context.Background()

	t.Run("it creates and returns a new user", func (t *testing.T) {
		user_service := &StubUserService{}
		auth_service := &StubAuthService{}
		org_service := &StubOrgService{}

		user_body := `
		{
			"email": "capybarasan@proton.me",
			"password": "#Capycapycapy890",
			"username": "capybara",
			"full_name" : "Capy Bara"
		}
		`
		body := bytes.NewReader([]byte(user_body))

		request, _ := http.NewRequest(http.MethodPost, "/api/auth/signup", body)
		response := httptest.NewRecorder()
		
		api_server := api.NewAPIServer(
			test_ctx, 
			user_service,
			auth_service,
			org_service,
		)

		api_server.ServeHTTP(response, request)

		got_status := response.Result().StatusCode
		want_status := http.StatusOK

		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}

		got_create_calls := user_service.create_n_calls
		want_create_calls := 1

		if got_create_calls != want_create_calls {
			t.Errorf("got user_service.Create called for %d times, want %d", got_create_calls, want_create_calls)
		}
	})

	t.Run("it returns status 422 on malformed request body", func (t *testing.T) {
		user_service := &StubUserService{}
		auth_service := &StubAuthService{}
		org_service := &StubOrgService{}

		api_server := api.NewAPIServer(
			test_ctx, 
			user_service,
			auth_service,
			org_service,
		)
		
		tests := []struct{
			desc string
			body any
		}{
			{"no body", nil},
			{"empty string", ""},
			{"invalid json 1", "{]}"},
			{"string", "huh"},
			{"invalid json 2", "{a: 123}"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 422 on %s", tt.desc), func (t *testing.T) {
				body_string, _ := tt.body.(string)
				body := bytes.NewReader([]byte(body_string))
				request, _ := http.NewRequest(http.MethodPost, "/api/auth/signup", body)
				response := httptest.NewRecorder()
				
				api_server.ServeHTTP(response, request)

				got_status := response.Result().StatusCode
				want_status := http.StatusUnprocessableEntity

				if got_status != want_status {
					t.Errorf("got status %d, want %d", got_status, want_status)
				}
			})
		}
	})

	t.Run("it returns status 400 and error messages on validation error", func (t *testing.T) {
		user_service := &StubUserService{}
		auth_service := &StubAuthService{}
		org_service := &StubOrgService{}

		api_server := api.NewAPIServer(
			test_ctx, 
			user_service,
			auth_service,
			org_service,
		)
		
		tests := []struct{
			desc string
			body any
			expected_errors []string
		}{
			{
				"incomplete payload", 
				`	
				{
					"email": "capybarasan@proton.me"
				}
				`,
				[]string{
					"passwords must have",
					"username must be",
					"full name",
				},
			},
			{
				"weak password", 
				`
				{
					"email": "capybarasan@proton.me",
					"password": "reallyweak",
					"username": "capybara",
					"full_name" : "Capy Bara"
				}
				`,
				[]string{
					"passwords must have at least",
				},
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				body_string, _ := tt.body.(string)
				body := bytes.NewReader([]byte(body_string))
				request, _ := http.NewRequest(http.MethodPost, "/api/auth/signup", body)
				response := httptest.NewRecorder()
				
				api_server.ServeHTTP(response, request)

				got_status := response.Result().StatusCode
				want_status := http.StatusBadRequest

				if got_status != want_status {
					t.Errorf("got status %d, want %d", got_status, want_status)
				}

				decoder := json.NewDecoder(response.Body)
				var response_body utils.BaseResponse[any]

				if err := decoder.Decode(&response_body); err != nil {
					t.Errorf("got response parsing err %v, want %v", err.Error(), nil)
				}

				got_error_msg := response_body.ErrorDetails.Message
				for _, want_error_msg := range tt.expected_errors {
					if !strings.Contains(got_error_msg, want_error_msg) {
						t.Errorf("got error message %s, want containing %s", got_error_msg, want_error_msg)
					}
				}

			})
		}
	})
}