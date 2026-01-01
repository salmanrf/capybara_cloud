package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/api/routes"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func TestCreateApplication(t *testing.T) {
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

		application_service.create_return = &database.Application{
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

		got_service_called := application_service.create_n_calls
		want_service_called := 1

		if got_service_called != want_service_called {
			t.Errorf("got application_service create method called %d times, want %d\n", got_service_called, want_service_called)
		}

		got_called_with_user_id := application_service.create_calls_arg1[0]
		want_called_with_user_id := expected_user_id

		if got_called_with_user_id != want_called_with_user_id {
			t.Errorf("got create called with user id %s, want %s", got_called_with_user_id, want_called_with_user_id)
		}
	})

	t.Run("should returns 403 error if doesn't have enough permission", func (t *testing.T) {
		defer func () {
			application_service.Clear()
		}()

		application_service.create_err = errors.New("permission_denied")
		
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
		application_service.update_return = &database.Application{
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
		application_service.update_return = &database.Application{
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

		got_update_called := application_service.update_n_calls
		want_update_callled := 1

		if got_update_called != want_update_callled {
			t.Errorf("got service method update called %d times, want %d\n", got_update_called, want_update_callled)
		}

		got_called_with_app_id := application_service.update_calls_arg1[0]
		want_called_with_app_id := expected_app_id

		if got_called_with_app_id != want_called_with_app_id {
			t.Errorf("got service method update called with app id %s, want %s\n", got_called_with_app_id, want_called_with_app_id)
		}

		got_called_with_user_id := application_service.update_calls_arg2[0]
		want_called_with_user_id := mock_user_id

		if got_called_with_user_id != want_called_with_user_id {
			t.Errorf("got service method update called with user id %s, want %s\n", got_called_with_user_id, want_called_with_user_id)
		}

		got_called_with_dto := application_service.update_calls_arg3[0]
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

		application_service.update_err = errors.New("permission_denied")
		
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

		application_service.find_one_return = nil
		application_service.find_one_error = nil
		
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

		application_service.find_one_return = nil
		application_service.find_one_error = errors.New("permission_denied")
		
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

				application_service.find_one_return = &database.FindOneApplicationWithProjectMemberRow{
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

				application_service.find_one_return = &database.FindOneApplicationWithProjectMemberRow{
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

				got_service_called_n_times := application_service.find_one_n_calls
				want_service_called_n_times := 1

				if got_service_called_n_times != want_service_called_n_times {
					t.Errorf("got service method called %d times, want %d", got_service_called_n_times, want_service_called_n_times)
				}
				
				got_service_called_with_app_id := application_service.find_one_calls_arg1[0]
				want_service_called_with_app_id := expected_app_id

				if got_service_called_with_app_id != want_service_called_with_app_id {
					t.Errorf("got service method called with app id %s, want %s", got_service_called_with_app_id, want_service_called_with_app_id)
				}
				
				got_service_called_with_user_id := application_service.find_one_calls_arg2[0]
				want_service_called_with_user_id := expected_user_id
				
				if got_service_called_with_user_id != want_service_called_with_user_id {
					t.Errorf("got service method called with user id %s, want %s", got_service_called_with_user_id, want_service_called_with_user_id)
				}
			})
		}
	})
}