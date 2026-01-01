package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/api/routes"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func TestCreateApplicationConfig(t *testing.T) {
	application_service := &StubApplicationService{}
	jwt_validator := &StubJwtValidator{}
	
	mux := chi.NewRouter()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(application_service, jwt_validator)) 

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

	t.Run("should returns status code 400 when validation failed", func (t *testing.T) {
		tests := []struct{
			desc string
			body any
		}{
			{
				"empty json",
				"{}",
			},
			{
				"empty variables_json",
				`
				{
					"foo": "bar"
				}
				`,
			},
			{
				"invalid variables_json 1",
				`
				{
					"variables_json": "123"
				}
				`,
			},
			{
				"invalid variables_json 2",
				`
				{
					"variables_json": "{foo: bar}"
				}
				`,
			},
			{
				"invalid variables_json 3",
				`
				{
					"variables_json": "foo: bar"
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
				res := httptest.NewRecorder()

				req.AddCookie(sid_cookie)

				api.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := http.StatusBadRequest

				if got_status != want_status {
					t.Errorf("got status code %d, want %d\n", got_status, want_status)
				}
			})
		}
	})

	t.Run("should returns status code 200 and the application config", func (t *testing.T) {
		varv_1 := "mongo://12345"
		variables_json_1 := fmt.Sprintf(`{\"MONGO_URI\": \"%s\"}`, varv_1)
		varv_21 := "abcd"
		varv_22 := "zxcv"
		variables_json_2 := fmt.Sprintf(`{\"AUTH0_CLIENT_ID\": \"%s\", \"AUTH0_CLIENT_SECRET\": \"%s\"}`, varv_21, varv_22)
		
		tests := []struct{
			desc string
			app_id string
			body any
			expected_variables_json string
			expected_config_variables map[string]any
		}{
			{
				"valid payload",
				"817c42f9-a216-4475-a6e5-d98864bb5161",
				fmt.Sprintf(`
				{
					"variables_json": "%s"
				}
				`, variables_json_1),
				variables_json_1,
				map[string]any{
					"MONGO_URI": varv_1, 
				},
			},
			{
				"valid payload",
				"eb29b17d-04c3-4895-a170-930c36766df7",
				fmt.Sprintf(`
				{
					"variables_json": "%s"
				}
				`, variables_json_2),
				variables_json_2,
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
				}
				
				application_service.create_config_return = expected_app_config
				
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

	t.Run("should call service method properly", func (t *testing.T) {
		tests := []struct{
			app_id string
			user_id string
			body string
		}{
			{
				"12345",
				"abcd",
				`
				{
					"variables_json": "{}"
				}
				`,
			},
			{
				"67890",
				"zxcv",
				`
				{
					"variables_json": "{}"
				}
				`,
			},
		}

		jwt_validator.validate_return = mock_user_id

		for _, tt := range tests {
			t.Run(fmt.Sprintf("app id %s, user id %s", tt.app_id, tt.user_id), func (t *testing.T) {
				defer func() {
					application_service.Clear()
				}()

				expected_app_config := &database.ApplicationConfig{}
				application_service.create_config_return = expected_app_config
				jwt_validator.validate_return = tt.user_id
				
				payload := tt.body

				expected_dto := dto.CreateApplicationConfigDto{
					VariablesJSON: "{}",
					VariablesMap: &map[string]any{},
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

				got_service_called_n_times := application_service.create_config_n_calls
				want_service_called_n_times := 1

				if got_service_called_n_times != want_service_called_n_times {
					t.Errorf("got service method called %d times, want %d", got_service_called_n_times, want_service_called_n_times)
				}

				got_service_called_with_app_id := application_service.create_config_calls_arg1[0]
				want_service_called_with_app_id := tt.app_id

				if got_service_called_with_app_id != want_service_called_with_app_id {
					t.Errorf("got service method called with app id %s, want %s", got_service_called_with_app_id, want_service_called_with_app_id)
				}

				got_service_called_with_user_id := application_service.create_config_calls_arg2[0]
				want_service_called_with_user_id := tt.user_id

				if got_service_called_with_user_id != want_service_called_with_user_id {
					t.Errorf("got service method called with user id %s, want %s", got_service_called_with_user_id, want_service_called_with_user_id)
				}

				got_service_called_with_dto := application_service.create_config_calls_arg3[0]
				want_service_called_with_dto := expected_dto

				if !reflect.DeepEqual(got_service_called_with_dto, want_service_called_with_dto) {
					t.Errorf("got service method called with dto %v, want %v", got_service_called_with_dto, want_service_called_with_dto)
				}
			})
		}
	})
}