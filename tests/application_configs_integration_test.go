package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/api/routes"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/shared/utils"
)

func TestCreateApplicationConfig(t *testing.T) {
	application_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}

	mux := chi.NewRouter()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(application_service, deployment_service, jwt_validator))

	type api_server struct {
		http.Handler
	}

	api := api_server{
		mux,
	}

	sid_cookie := &http.Cookie{
		Name: "sid",
		Value: "123",
		Path: "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge: 3600 * 24,
		HttpOnly: true,
		Secure: os.Getenv("STAGE") != "local",
	}

	mock_user_id := "9ae9a0b2-d09e-4dcf-a0b1-18316fcef6cc"

	t.Run("should return status code 401 if not logged in", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"

		req, _ := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			nil,
		)
		res := httptest.NewRecorder()

		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusUnauthorized

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}
	})

	t.Run("should return status code 400/422 when validation failed", func (t *testing.T) {
		tests := []struct{
			desc string
			body any
		}{
			{
				"empty json",
				"{}",
			},
			{
				"empty variables",
				`
				{
					"foo": "bar"
				}
				`,
			},
			{
				"invalid variables 1",
				`
				{
					"variables": "123"
				}
				`,
			},
			{
				"invalid variables",
				`
				{
					"variables": "{foo: bar}"
				}
				`,
			},
			{
				"invalid variables",
				`
				{
					"variables": "foo: bar"
				}
				`,
			},
			{
				"invalid variables (non primitive data type)",
				`
				{
					"variables": {
						"foo": {
							"bar": "baz"
						}
					}
				}
				`,
			},
			{
				"invalid variables (missing port)",
				`
				{
					"variables": {
						"POSTGRES_URI": "postgres://abcd:abcd@zsh:3220"
					}
				}
				`,
			},
		}

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id

		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				defer func() {
				application_service.Clear()
			}()
				
				payload, _ := tt.body.(string)
				req_body := bytes.NewBuffer([]byte(payload))
				req, _ := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
					req_body,
				)
				req.AddCookie(sid_cookie)
				
				res := httptest.NewRecorder()
				api.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := []int{http.StatusBadRequest, http.StatusUnprocessableEntity}
				if slices.Index(want_status, got_status) == -1 {
					t.Errorf("got status code %d, want %v\n", got_status, want_status)
				}
			})
		}
	})

	t.Run("should return status code 404 on not_found errors", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()
		
		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id
		application_service.Create_config_err = errors.New("not_found")

		body_string := `
			{
				"port": 2222,
				"variables": {
					"foo": "bar"
				}
			}
		`
		req_body := bytes.NewBuffer([]byte(body_string))
		req, _ := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			req_body,
		)
		req.AddCookie(sid_cookie)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound
		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("should return status code 403 on permission_denied errors", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()
		
		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id
		application_service.Create_config_err = errors.New("permission_denied")

		body_string := `
			{
				"port": 1111,
				"variables": {
					"foo": "bar"
				}
			}
		`
		req_body := bytes.NewBuffer([]byte(body_string))
		req, _ := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			req_body,
		)
		req.AddCookie(sid_cookie)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusForbidden
		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})
	
	t.Run("should return status code 500 on other errors", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()
		
		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id
		application_service.Create_config_err = errors.New("internal server error")

		body_string := `
			{
				"port": 7777, 
				"variables": {
					"foo": "bar"
				}
			}
		`
		req_body := bytes.NewBuffer([]byte(body_string))
		req, _ := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			req_body,
		)
		req.AddCookie(sid_cookie)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusInternalServerError
		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("should return status code 200 and the application config", func (t *testing.T) {
		varv_1 := "mongo://12345"
		variables_1 := fmt.Sprintf(`{"MONGO_URI": "%s"}`, varv_1)
		varv_21 := "abcd"
		varv_22 := "zxcv"
		variables_2 := fmt.Sprintf(`{"AUTH0_CLIENT_ID": "%s", "AUTH0_CLIENT_SECRET": "%s"}`, varv_21, varv_22)
		
		tests := []struct{
			desc string
			app_id string
			body any
			expected_port int
			expected_variables_json string
			expected_config_variables map[string]any
		}{
			{
				"valid payload",
				"817c42f9-a216-4475-a6e5-d98864bb5161",
				fmt.Sprintf(`
				{
					"port": 8080,
					"variables": %s
				}
				`, variables_1),
				8080,
				variables_1,
				map[string]any{
					"MONGO_URI": varv_1, 
				},
			},
			{
				"valid payload",
				"eb29b17d-04c3-4895-a170-930c36766df7",
				fmt.Sprintf(`
				{
					"port": 5555,
					"variables": %s
				}
				`, variables_2),
				5555,
				variables_2,
				map[string]any{
					"AUTH0_CLIENT_ID": varv_21,
					"AUTH0_CLIENT_SECRET": varv_22,
				},
			},
		}

		jwt_validator.validate_return = mock_user_id

		for i, tt := range tests {
			t.Run(fmt.Sprintf("returns 200 on %s (%d)", tt.desc, i + 1), func (t *testing.T) {
				defer func() {
					application_service.Clear()
				}()

				app_uuid := pgtype.UUID{}
				app_uuid.Scan(tt.app_id)
				app_cfg_uuid := pgtype.UUID{}
				expected_app_config := &database.ApplicationConfig{
					AppID: app_uuid,
					AppCfgID: app_cfg_uuid,
					VariablesJson: []byte(tt.expected_variables_json),
					Port: int32(tt.expected_port),
				}
				
				application_service.Create_config_return = expected_app_config
				
				payload, _ := tt.body.(string)

				req_body := bytes.NewBuffer([]byte(payload))
				req, _ := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("/api/applications/%s/configs", tt.app_id),
					req_body,
				)
				res := httptest.NewRecorder()

				req.AddCookie(sid_cookie)

				api.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := http.StatusOK

				if got_status != want_status {
					t.Errorf("got status code %d, want %d\n", got_status, want_status)
				}

				decoder := json.NewDecoder(res.Result().Body)
				var got_body utils.BaseResponse[map[string]any]
				if err := decoder.Decode(&got_body); err != nil {
					t.Errorf("got error parsing body %v, want nil", err)
				}

				got_call_arg := application_service.Create_config_calls_arg3[0]
				got_called_port := got_call_arg.Port

				got_data, _ := got_body.Data.(map[string]any)				

				want_called_port := tt.expected_port
				if got_called_port != want_called_port {
					t.Errorf("got port %v, want %v", got_called_port, want_called_port)
				}
				
				got_config_variables, ok := got_data["config_variables"].(map[string]any)
				want_config_variables := tt.expected_config_variables
				if !ok || !reflect.DeepEqual(got_config_variables, want_config_variables) {
					t.Errorf("got config variables %v, want %v", got_config_variables, want_config_variables)
				}

				got_variables_json, ok := got_data["variables_json"].(string)
				want_variables_json := string(expected_app_config.VariablesJson)

				if !ok || strings.Compare(got_variables_json, want_variables_json) != 0 {
					t.Errorf("got variables json %s, want %s", got_variables_json, want_variables_json)
				}
			})
		}
	})

	t.Run("should call service method properly", func (t *testing.T) {
		tests := []struct{
			app_id string
			user_id string
			body string
			expected_port int
			expected_map map[string]any
		}{
			{
				"12345",
				"abcd",
				`
				{
					"port": 1025,
					"variables": {
						"A": "B"
					}
				}
				`,
				1025,
				map[string]any{
					"A": "B",
				},
			},
			{
				"67890",
				"zxcv",
				`
				{
					"port": 8888,
					"variables": {
						"C": "D"
					}
				}
				`,
				8888,
				map[string]any{
					"C": "D",
				},
			},
		}

		jwt_validator.validate_return = mock_user_id

		for _, tt := range tests {
			t.Run(fmt.Sprintf("app id %s, user id %s", tt.app_id, tt.user_id), func (t *testing.T) {
				defer func() {
					application_service.Clear()
				}()

				expected_app_config := &database.ApplicationConfig{}
				application_service.Create_config_return = expected_app_config
				jwt_validator.validate_return = tt.user_id
				
				payload := tt.body

				expected_dto := dto.CreateApplicationConfigDto{
					Port: tt.expected_port,
					Variables: tt.expected_map,
				}

				req_body := bytes.NewBuffer([]byte(payload))
				req, _ := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("/api/applications/%s/configs", tt.app_id),
					req_body,
				)
				res := httptest.NewRecorder()

				req.AddCookie(sid_cookie)

				api.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := http.StatusOK

				if got_status != want_status {
					t.Errorf("got status code %d, want %d\n", got_status, want_status)
				}

				got_service_called_n_times := application_service.Create_config_n_calls
				want_service_called_n_times := 1

				if got_service_called_n_times != want_service_called_n_times {
					t.Errorf("got service method called %d times, want %d", got_service_called_n_times, want_service_called_n_times)
				}

				got_service_called_with_app_id := application_service.Create_config_calls_arg1[0]
				want_service_called_with_app_id := tt.app_id

				if got_service_called_with_app_id != want_service_called_with_app_id {
					t.Errorf("got service method called with app id %s, want %s", got_service_called_with_app_id, want_service_called_with_app_id)
				}

				got_service_called_with_user_id := application_service.Create_config_calls_arg2[0]
				want_service_called_with_user_id := tt.user_id

				if got_service_called_with_user_id != want_service_called_with_user_id {
					t.Errorf("got service method called with user id %s, want %s", got_service_called_with_user_id, want_service_called_with_user_id)
				}

				got_service_called_with_dto := application_service.Create_config_calls_arg3[0]
				want_service_called_with_dto := expected_dto

				if !reflect.DeepEqual(got_service_called_with_dto, want_service_called_with_dto) {
					t.Errorf("got service method called with dto %v, want %v", got_service_called_with_dto, want_service_called_with_dto)
				}
			})
		}
	})
}

func TestFindOneApplicationConfig(t *testing.T) {
	application_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}

	mux := chi.NewRouter()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(application_service, deployment_service, jwt_validator))

	type api_server struct {
		http.Handler
	}

	api := api_server{
		mux,
	}

	sid_cookie := &http.Cookie{
		Name: "sid",
		Value: "123",
		Path: "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge: 3600 * 24,
		HttpOnly: true,
		Secure: os.Getenv("STAGE") != "local",
	}

	mock_user_id := "9ae9a0b2-d09e-4dcf-a0b1-18316fcef6cc"

	t.Run("should return status code 401 if not logged in", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"

		req, _ := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			nil,
		)
		res := httptest.NewRecorder()

		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusUnauthorized

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}
	})

	t.Run("should return status code 404 on not_found errors", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id
		application_service.Find_one_config_error = errors.New("not_found")

		req, _ := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			nil,
		)
		req.AddCookie(sid_cookie)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound
		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("should return status code 403 on permission_denied errors", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id
		application_service.Find_one_config_error = errors.New("permission_denied")

		req, _ := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			nil,
		)
		req.AddCookie(sid_cookie)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusForbidden
		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("should return status code 500 on other errors", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		jwt_validator.validate_return = mock_user_id
		application_service.Find_one_config_error = errors.New("internal server error")

		req, _ := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/applications/%s/configs", expected_app_id),
			nil,
		)
		req.AddCookie(sid_cookie)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusInternalServerError
		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("should return status code 200 and the application config", func (t *testing.T) {
		varv_1 := "mongo://12345"
		variables_1 := fmt.Sprintf(`{"MONGO_URI": "%s"}`, varv_1)
		varv_21 := "abcd"
		varv_22 := "zxcv"
		variables_2 := fmt.Sprintf(`{"AUTH0_CLIENT_ID": "%s", "AUTH0_CLIENT_SECRET": "%s"}`, varv_21, varv_22)
		
		tests := []struct{
			app_id string
			expected_variables_json string
			expected_config_variables map[string]any
		}{
			{
				"817c42f9-a216-4475-a6e5-d98864bb5161",
				variables_1,
				map[string]any{
					"MONGO_URI": varv_1, 
				},
			},
			{
				"eb29b17d-04c3-4895-a170-930c36766df7",
				variables_2,
				map[string]any{
					"AUTH0_CLIENT_ID": varv_21,
					"AUTH0_CLIENT_SECRET": varv_22,
				},
			},
		}

		jwt_validator.validate_return = mock_user_id

		for i, tt := range tests {
			t.Run(fmt.Sprintf("returns 200 (%d)", i + 1), func (t *testing.T) {
				defer func() {
					application_service.Clear()
				}()

				app_uuid := pgtype.UUID{}
				app_uuid.Scan(tt.app_id)
				expected_app_config := &dto.ApplicationConfigResponse{
					AppID: tt.app_id,
					AppCfgID: tt.app_id,
					Port: 8080,
					VariablesJson: tt.expected_variables_json,
					ConfigVariables: tt.expected_config_variables,
				}
				
				application_service.Find_one_config_return = expected_app_config
				
				req, _ := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("/api/applications/%s/configs", tt.app_id),
					nil,
				)
				res := httptest.NewRecorder()

				req.AddCookie(sid_cookie)

				api.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := http.StatusOK

				if got_status != want_status {
					t.Errorf("got status code %d, want %d\n", got_status, want_status)
				}

				decoder := json.NewDecoder(res.Result().Body)
				var got_body utils.BaseResponse[map[string]any]
				if err := decoder.Decode(&got_body); err != nil {
					t.Errorf("got error parsing body %v, want nil", err)
				}

				got_data, _ := got_body.Data.(map[string]any)				
				got_config_variables, ok := got_data["config_variables"].(map[string]any)
				want_config_variables := tt.expected_config_variables
				if !ok || !reflect.DeepEqual(got_config_variables, want_config_variables) {
					t.Errorf("got config variables %v, want %v", got_config_variables, want_config_variables)
				}

				got_variables_json, ok := got_data["variables_json"].(string)
				want_variables_json := string(expected_app_config.VariablesJson)

				if !ok || strings.Compare(got_variables_json, want_variables_json) != 0 {
					t.Errorf("got variables json %s, want %s", got_variables_json, want_variables_json)
				}
			})
		}
	})
}