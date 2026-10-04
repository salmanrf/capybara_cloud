package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"maps"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/api/routes"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/logger"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

func TestCreateApplication(t *testing.T) {
	application_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}

	mux := chi.NewRouter()
	logger, cleanup, _ := logger.InitLogger("", nil)
	defer cleanup()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(logger, application_service, deployment_service, jwt_validator))

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
	
	t.Run("should returns status code 201 and the new application on success", func (t *testing.T) {
		defer func () {
			application_service.Clear()
		}()
		
		expected_app_uuid := pgtype.UUID{}
		expected_app_uuid.Scan("a689caa1-6cdb-4d2c-8db0-5a30ee6f83fa")
		expected_type := dto.GetSupportedAppTypes()[0]
		expected_project_uuid := pgtype.UUID{}
		expected_project_uuid.Scan("28451bd5-0113-4ec6-9540-6646ae72a957")
		expected_app_name := "Sophia School"

		application_service.Create_return = &database.Application{
			AppID: expected_app_uuid,
			ProjectID: expected_project_uuid,
			Name: expected_app_name,
			Type: expected_type,
		}

		req_body := bytes.NewBuffer([]byte(
			fmt.Sprintf(
				`
					{
						"project_id": "%s",
						"type": "%s",
						"name": "%s"
					}
				`,
				expected_project_uuid.String(),
				expected_type,
				expected_app_name,
			),
		))
		req, _ := http.NewRequest(http.MethodPost, "/api/applications", req_body)
		res := httptest.NewRecorder()
		req.AddCookie(sid_cookie)

		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusCreated

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}

		decoder := json.NewDecoder(res.Result().Body)
		var got_body map[string]any
		if err := decoder.Decode(&got_body); err != nil {
			t.Errorf("got error parsing response body %v, want nil\n", err)
		}

		got_data, ok := got_body["data"].(map[string]any)
		if !ok {
			t.Errorf("got %q, want data\n", got_data)
		}

		got_app_id, ok := got_data["app_id"].(string)
		want_app_id := expected_app_uuid.String()
		if !ok || got_app_id != want_app_id {
			t.Errorf("got app id %v, want app id %v\n", got_app_id, want_app_id)
		}

		got_app_name, ok := got_data["name"].(string)
		want_app_name := expected_app_name
		if !ok || got_app_name != want_app_name {
			t.Errorf("got app name %v, want app name %v\n", got_app_name, want_app_name)
		}
		got_project_id, ok := got_data["project_id"].(string)
		want_project_id := expected_project_uuid.String()
		if !ok || got_project_id != want_project_id {
			t.Errorf("got project_id %v, want project_id %v\n", got_project_id, want_project_id)
		}
	})

	t.Run("should create application on behalf of logged in user", func (t *testing.T) {
		defer func () {
			application_service.Clear()
		}()
		
		expected_user_id := "123"
		expected_type := dto.GetSupportedAppTypes()[0]
		expected_project_uuid := pgtype.UUID{}
		expected_project_uuid.Scan("28451bd5-0113-4ec6-9540-6646ae72a957")
		expected_app_name := "Sophia School"

		jwt_validator.validate_return = expected_user_id

		req_body := bytes.NewBuffer([]byte(
			fmt.Sprintf(
				`
					{
						"project_id": "%s",
						"type": "%s",
						"name": "%s"
					}
				`,
				expected_project_uuid.String(),
				expected_type,
				expected_app_name,
			),
		))
		req, _ := http.NewRequest(http.MethodPost, "/api/applications", req_body)
		res := httptest.NewRecorder()
		req.AddCookie(sid_cookie)

		api.ServeHTTP(res, req)

		got_service_called := application_service.Create_n_calls
		want_service_called := 1

		if got_service_called != want_service_called {
			t.Errorf("got application_service create method called %d times, want %d\n", got_service_called, want_service_called)
		}

		got_called_with_user_id := application_service.Create_calls_arg1[0]
		want_called_with_user_id := expected_user_id

		if got_called_with_user_id != want_called_with_user_id {
			t.Errorf("got create called with user id %s, want %s", got_called_with_user_id, want_called_with_user_id)
		}
	})

	t.Run("should returns 403 error if doesn't have enough permission", func (t *testing.T) {
		defer func () {
			application_service.Clear()
		}()

		application_service.Create_err = errors.New("permission_denied")
		
		expected_user_id := "123"
		expected_type := dto.GetSupportedAppTypes()[0]
		expected_project_uuid := pgtype.UUID{}
		expected_project_uuid.Scan("28451bd5-0113-4ec6-9540-6646ae72a957")
		expected_app_name := "Sophia School"

		jwt_validator.validate_return = expected_user_id

		req_body := bytes.NewBuffer([]byte(
			fmt.Sprintf(
				`
					{
						"project_id": "%s",
						"type": "%s",
						"name": "%s"
					}
				`,
				expected_project_uuid.String(),
				expected_type,
				expected_app_name,
			),
		))
		req, _ := http.NewRequest(http.MethodPost, "/api/applications", req_body)
		res := httptest.NewRecorder()
		req.AddCookie(sid_cookie)

		api.ServeHTTP(res, req)

		got_status_code := res.Result().StatusCode
		want_status_code := http.StatusForbidden

		if got_status_code != want_status_code {
			t.Errorf("got status code %d, want %d\n", got_status_code, want_status_code)
		}
	})

	t.Run("should return status code 401 if not logged in", func (t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/applications", nil)
		res := httptest.NewRecorder()

		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusUnauthorized

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}
	})

	t.Run("should returns status code 422 when received malformed payload", func (t *testing.T) {
		tests := []struct{
			desc string
			body any
		}{
			{
				"malformed json",
				`
				{
					foobar: baz
				}
				`,
			},
			{
				"nil",
				nil,
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 422 on %s", tt.desc), func (t *testing.T) {
				defer func () {
					application_service.Clear()
				}()
				
				payload, _ := tt.body.(string)
				req_body := bytes.NewBuffer([]byte(payload))
				req, _ := http.NewRequest(http.MethodPost, "/api/applications", req_body)
				res := httptest.NewRecorder()
				req.AddCookie(sid_cookie)

				api.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := http.StatusUnprocessableEntity

				if got_status != want_status {
					t.Errorf("got status code %d, want %d\n", got_status, want_status)
				}
			})
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
				"empty name",
				`
				{
					"foo": "bar"
				}
				`,
			},
			{
				"empty project_id",
				`
				{
					"name": "Ada Computer Hardwares"
				}
				`,
			},
			{
				"unsupported app type",
				`
				{
					"project_id": "28451bd5-0113-4ec6-9540-6646ae72a957",
					"name": "Ada Computer Hardwares",
					"type": "native_compute_intensive"
				}
				`,
			},
			{
				"invalid project_id",
				`
				{
					"project_id": "28451bd5-0113-4ec6",
					"name": "Ada Computer Hardwares",
					"type": "native_compute_intensive"
				}
				`,
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				defer func () {
					application_service.Clear()
				}()
				
				payload, _ := tt.body.(string)
				req_body := bytes.NewBuffer([]byte(payload))
				req, _ := http.NewRequest(http.MethodPost, "/api/applications", req_body)
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
}

func TestUpdateApplication(t *testing.T) {
	application_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}

	mux := chi.NewRouter()
	logger, cleanup, _ := logger.InitLogger("", nil)
	defer cleanup()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(logger, application_service, deployment_service, jwt_validator))

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
			http.MethodPut, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
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

	t.Run("should return status code 200 on success and the updated application", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		expected_new_name := "Ada Hardware 2"
		req_body := bytes.NewBuffer([]byte(
			// ? ignore and check assert to whatever is returned by the service
			`
				{
					"name": "12345" 
				}
			`,
		))

		jwt_validator.validate_return = mock_user_id
		application_service.Update_return = &database.Application{
			Name: expected_new_name,
		}

		req, _ := http.NewRequest(
			http.MethodPut, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
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
			t.Errorf("got error parsing response body %v, want nil", err)
		}

		data, ok := got_body.Data.(map[string]any)
		got_body_name := data["name"].(string)
		want_name := expected_new_name
		if !ok || got_body_name != want_name {
			t.Errorf("got new app name %s, want %s", got_body_name, want_name)
		}
	})

	t.Run("should call service method correctly", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		expected_new_name := "Ada Hardware"
		req_body := bytes.NewBuffer([]byte(
			fmt.Sprintf(
				`
					{
						"name": "%s"
					}
				`,
				expected_new_name,
			),
		))

		jwt_validator.validate_return = mock_user_id
		application_service.Update_return = &database.Application{
			Name: expected_new_name,
		}

		req, _ := http.NewRequest(
			http.MethodPut, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
			req_body,
		)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		api.ServeHTTP(res, req)

		got_update_called := application_service.Update_n_calls
		want_update_callled := 1

		if got_update_called != want_update_callled {
			t.Errorf("got service method update called %d times, want %d\n", got_update_called, want_update_callled)
		}

		got_called_with_app_id := application_service.Update_calls_arg1[0]
		want_called_with_app_id := expected_app_id

		if got_called_with_app_id != want_called_with_app_id {
			t.Errorf("got service method update called with app id %s, want %s\n", got_called_with_app_id, want_called_with_app_id)
		}

		got_called_with_user_id := application_service.Update_calls_arg2[0]
		want_called_with_user_id := mock_user_id

		if got_called_with_user_id != want_called_with_user_id {
			t.Errorf("got service method update called with user id %s, want %s\n", got_called_with_user_id, want_called_with_user_id)
		}

		got_called_with_dto := application_service.Update_calls_arg3[0]
		want_called_with_dto := dto.UpdateApplicationDto{
			Name: expected_new_name,
		}

		if got_called_with_dto != want_called_with_dto {
			t.Errorf("got service method update called with dto %v, want %v\n", got_called_with_dto, want_called_with_dto)
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
				"empty name",
				`
				{
					"foo": "bar"
				}
				`,
			},
		}

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		
		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				payload, _ := tt.body.(string)
				req_body := bytes.NewBuffer([]byte(payload))
				req, _ := http.NewRequest(
					http.MethodPut, 
					fmt.Sprintf("/api/applications/%s", expected_app_id), 
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

	t.Run("should return status code 403 if doesn't have sufficient permission", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		expected_new_name := "Ada Hardware"
		req_body := bytes.NewBuffer([]byte(
			fmt.Sprintf(
				`
					{
						"name": "%s"
					}
				`,
				expected_new_name,
			),
		))

		application_service.Update_err = errors.New("permission_denied")
		
		req, _ := http.NewRequest(
			http.MethodPut, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
			req_body,
		)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusForbidden

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}
	})
}

func TestFindOneApplication(t *testing.T) {
	application_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}

	mux := chi.NewRouter()
	logger, cleanup, _ := logger.InitLogger("", nil)
	defer cleanup()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(logger, application_service, deployment_service, jwt_validator))

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

	t.Run("should return status code 401 if not logged in", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		
		req, _ := http.NewRequest(
			http.MethodGet, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
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

	t.Run("should return status code 404 if server returns nil", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"

		application_service.Find_one_return = nil
		application_service.Find_one_error = nil
		
		req, _ := http.NewRequest(
			http.MethodGet, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
			nil,
		)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)
		
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}
	})

	t.Run("should return status code 403 doesn't have enough permission", func (t *testing.T) {
		defer func() {
			application_service.Clear()
		}()

		expected_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"

		application_service.Find_one_return = nil
		application_service.Find_one_error = errors.New("permission_denied")
		
		req, _ := http.NewRequest(
			http.MethodGet, 
			fmt.Sprintf("/api/applications/%s", expected_app_id), 
			nil,
		)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)
		
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusForbidden

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}
	})
	
	t.Run("should return the application from service", func (t *testing.T) {
		tests := []struct{
			app_id string
		}{
			{"7aaa1bf8-437f-4f3c-8691-8316fc6fbeaa"},
			{"7aaa1bf8-437f-4f3c-8691-8316fc6fbebb"},
			{"7aaa1bf8-437f-4f3c-8691-8316fc6fbecc"},
		}

		for i, tt := range tests {
			t.Run(fmt.Sprintf("%d", i), func (t *testing.T) {
				defer func() {
					application_service.Clear()
				}()
				
				expected_app_id := tt.app_id
				expected_app_uuid := pgtype.UUID{}
				expected_app_uuid.Scan(expected_app_id)

				application_service.Find_one_return = &database.FindOneApplicationCompleteRow{
					AppID: expected_app_uuid,
				}
				
				req, _ := http.NewRequest(
					http.MethodGet, 
					fmt.Sprintf("/api/applications/%s", expected_app_id), 
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
				var got_body utils.BaseResponse[any]
				if err := decoder.Decode(&got_body); err != nil {
					t.Errorf("got error parsing response body %v, want nil", err)
				}

				got_data, ok := got_body.Data.(map[string]any) 
				if !ok {
					t.Errorf("got response data %v, want map", got_data)
				}

				got_app_id, ok := got_data["app_id"].(string) 
				want_app_id := expected_app_id
				if !ok || got_app_id != expected_app_id {
					t.Errorf("got app_id %s, want %s", got_app_id, want_app_id)
				}
			})
		}
	})

	t.Run("should call service method properly", func (t *testing.T) {
		tests := []struct{
			app_id string
		}{
			{"7aaa1bf8-437f-4f3c-8691-8316fc6fbeaa"},
			{"7aaa1bf8-437f-4f3c-8691-8316fc6fbebb"},
			{"7aaa1bf8-437f-4f3c-8691-8316fc6fbecc"},
		}

		expected_user_id := "abcd"
		jwt_validator.validate_return = expected_user_id

		for i, tt := range tests {
			t.Run(fmt.Sprintf("%d", i), func (t *testing.T) {
				defer func() {
					application_service.Clear()
				}()
				
				expected_app_id := tt.app_id
				expected_app_uuid := pgtype.UUID{}
				expected_app_uuid.Scan(expected_app_id)

				application_service.Find_one_return = &database.FindOneApplicationCompleteRow{
					AppID: expected_app_uuid,
				}
				
				req, _ := http.NewRequest(
					http.MethodGet, 
					fmt.Sprintf("/api/applications/%s", expected_app_id), 
					nil,
				)
				res := httptest.NewRecorder()

				req.AddCookie(sid_cookie)

				api.ServeHTTP(res, req)

				got_service_called_n_times := application_service.Find_one_n_calls
				want_service_called_n_times := 1

				if got_service_called_n_times != want_service_called_n_times {
					t.Errorf("got service method called %d times, want %d", got_service_called_n_times, want_service_called_n_times)
				}
				
				got_service_called_with_app_id := application_service.Find_one_calls_arg1[0]
				want_service_called_with_app_id := expected_app_id

				if got_service_called_with_app_id != want_service_called_with_app_id {
					t.Errorf("got service method called with app id %s, want %s", got_service_called_with_app_id, want_service_called_with_app_id)
				}
				
				got_service_called_with_user_id := application_service.Find_one_calls_arg2[0]
				want_service_called_with_user_id := expected_user_id
				
				if got_service_called_with_user_id != want_service_called_with_user_id {
					t.Errorf("got service method called with user id %s, want %s", got_service_called_with_user_id, want_service_called_with_user_id)
				}
			})
		}
	})
}
func TestListApplications(t *testing.T) {
	application_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}

	mock_user_id := "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"
	jwt_validator.validate_return = mock_user_id

	mux := chi.NewRouter()
	logger, cleanup, _ := logger.InitLogger("", nil)
	defer cleanup()
	mux.Mount("/api/applications", routes.SetupApplicationRouter(logger, application_service, deployment_service, jwt_validator))

	sid_cookie := &http.Cookie{
		Name: "sid",
		Value: "123",
		Path: "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge: 3600 * 24,
		HttpOnly: true,
		Secure: os.Getenv("STAGE") != "local",
	}

	mock_project_id := "64c5e7da-3e02-4db8-aa2a-aa5161c085f7"

	tests := []struct{
		name string
		url string
		with_sid bool
		service_err error
		want_status int
		want_n_calls int
	}{
		{
			name: "it should return status 401 when not logged in",
			url: fmt.Sprintf("/api/applications?project_id=%s", mock_project_id),
			with_sid: false,
			want_status: http.StatusUnauthorized,
			want_n_calls: 0,
		},
		{
			name: "it should return status 400 when project_id is missing",
			url: "/api/applications",
			with_sid: true,
			want_status: http.StatusBadRequest,
			want_n_calls: 0,
		},
		{
			name: "it should return status 400 when project_id is empty",
			url: "/api/applications?project_id=",
			with_sid: true,
			want_status: http.StatusBadRequest,
			want_n_calls: 0,
		},
		{
			name: "it should return status 400 when project_id is not a uuid",
			url: "/api/applications?project_id=not-a-uuid",
			with_sid: true,
			want_status: http.StatusBadRequest,
			want_n_calls: 0,
		},
		{
			name: "it should return status 200 when project_id is valid",
			url: fmt.Sprintf("/api/applications?project_id=%s", mock_project_id),
			with_sid: true,
			want_status: http.StatusOK,
			want_n_calls: 1,
		},
		{
			name: "it should return status 500 when the service fails",
			url: fmt.Sprintf("/api/applications?project_id=%s", mock_project_id),
			with_sid: true,
			service_err: errors.New("unable to list applications, db query failed"),
			want_status: http.StatusInternalServerError,
			want_n_calls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func (t *testing.T) {
			defer application_service.Clear()

			application_service.List_by_project_err = tc.service_err

			req, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			if tc.with_sid {
				req.AddCookie(sid_cookie)
			}
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			got_status := res.Result().StatusCode
			if got_status != tc.want_status {
				t.Errorf("got status %d, want %d", got_status, tc.want_status)
			}

			got_n_calls := application_service.List_by_project_n_calls
			if got_n_calls != tc.want_n_calls {
				t.Errorf("got list by project called %d times, want %d", got_n_calls, tc.want_n_calls)
			}
		})
	}

	t.Run("it should pass the caller's user_id and the project_id to the service", func (t *testing.T) {
		defer application_service.Clear()

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/applications?project_id=%s", mock_project_id), nil)
		req.AddCookie(sid_cookie)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		if application_service.List_by_project_n_calls != 1 {
			t.Fatalf("got list by project called %d times, want 1", application_service.List_by_project_n_calls)
		}

		got_user_id := application_service.List_by_project_calls_arg1[0]
		if got_user_id != mock_user_id {
			t.Errorf("got user_id %s, want %s", got_user_id, mock_user_id)
		}

		got_project_id := application_service.List_by_project_calls_arg2[0]
		if got_project_id != mock_project_id {
			t.Errorf("got project_id %s, want %s", got_project_id, mock_project_id)
		}
	})

	t.Run("it should return the service's entries in the response body", func (t *testing.T) {
		defer application_service.Clear()

		mock_time := time.Date(2026, 9, 3, 8, 0, 0, 0, time.UTC)
		application_service.List_by_project_return = []dto.ApplicationListEntry{
			{
				AppID: "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50",
				Name: "api",
				Type: "container_nodejs",
				CreatedAt: mock_time,
				UpdatedAt: mock_time,
				Status: "running",
				LatestDeployment: &dto.ApplicationListLatestDeployment{
					AppDpID: "c0a1f1c6-2f7e-4c4b-9a39-0d3a1b2c3d4e",
					VersionNumber: 4,
					Status: -1,
					Outcome: "failed",
					CreatedAt: mock_time,
					UpdatedAt: mock_time,
				},
			},
			{
				AppID: "7aaa1bf8-437f-4f3c-8691-8316fc6fbe51",
				Name: "worker",
				Type: "container_nodejs",
				CreatedAt: mock_time,
				UpdatedAt: mock_time,
				Status: "not_deployed",
				LatestDeployment: nil,
			},
		}

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/applications?project_id=%s", mock_project_id), nil)
		req.AddCookie(sid_cookie)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		var body struct {
			Data []map[string]any `json:"data"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("unable to decode response body: %v", err)
		}
		if len(body.Data) != 2 {
			t.Fatalf("got %d entries, want 2", len(body.Data))
		}

		got_entry_keys := slices.Sorted(maps.Keys(body.Data[0]))
		want_entry_keys := []string{"app_id", "created_at", "latest_deployment", "name", "status", "type", "updated_at"}
		if !slices.Equal(got_entry_keys, want_entry_keys) {
			t.Errorf("got entry keys %v, want %v", got_entry_keys, want_entry_keys)
		}

		got_status := body.Data[0]["status"]
		if got_status != "running" {
			t.Errorf("got status %v, want running", got_status)
		}

		got_latest, ok := body.Data[0]["latest_deployment"].(map[string]any)
		if !ok {
			t.Fatalf("got latest_deployment %v, want an object", body.Data[0]["latest_deployment"])
		}

		got_latest_keys := slices.Sorted(maps.Keys(got_latest))
		want_latest_keys := []string{"app_dp_id", "created_at", "outcome", "status", "updated_at", "version_number"}
		if !slices.Equal(got_latest_keys, want_latest_keys) {
			t.Errorf("got latest_deployment keys %v, want %v", got_latest_keys, want_latest_keys)
		}
		if got_latest["outcome"] != "failed" {
			t.Errorf("got outcome %v, want failed", got_latest["outcome"])
		}
		if got_latest["status"] != float64(-1) {
			t.Errorf("got raw status %v, want -1", got_latest["status"])
		}

		got_second_status := body.Data[1]["status"]
		if got_second_status != "not_deployed" {
			t.Errorf("got status %v, want not_deployed", got_second_status)
		}

		got_second_latest, present := body.Data[1]["latest_deployment"]
		if !present {
			t.Fatalf("got no latest_deployment key, want null")
		}
		if got_second_latest != nil {
			t.Errorf("got latest_deployment %v, want null", got_second_latest)
		}
	})

	t.Run("it should return an empty list when the service returns no entries", func (t *testing.T) {
		defer application_service.Clear()

		application_service.List_by_project_return = []dto.ApplicationListEntry{}

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/applications?project_id=%s", mock_project_id), nil)
		req.AddCookie(sid_cookie)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		var body struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("unable to decode response body: %v", err)
		}

		got_data := string(body.Data)
		want_data := "[]"
		if got_data != want_data {
			t.Errorf("got data %s, want %s", got_data, want_data)
		}
	})
}
