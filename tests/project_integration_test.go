package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/api"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func TestProjectCreateIntegration(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
	project_service := &StubProjectService{}
	jwt_validator := &StubJwtValidator{}

	mock_user := &database.User{
		Email: "Salman",
		UserID: pgtype.UUID{},
		Username: "frnamlas",
		CreatedAt: pgtype.Timestamp{},
		UpdatedAt: pgtype.Timestamp{},
		FullName: "Salman RF",
	}
	user_service.find_by_id_return = mock_user
	user_service.find_by_id_err = nil

	server := api.NewAPIServer(
		test_ctx,
		user_service,
		auth_service,
		org_service,
		project_service,
		jwt_validator,
	)
	
	t.Run("it returns status code 201 and the new project on success", func (t *testing.T) {
		user_id := "123"
		org_id := "64c5e7da-3e02-4db8-aa2a-aa5161c085f7"
		project_name := "Capybara Org" 
		req_body := bytes.NewReader([]byte(fmt.Sprintf(`
			{
				"org_id": "%s",
				"name": "%s"
			}
		`, org_id, project_name)))
		req, _ := http.NewRequest(http.MethodPost, "/api/projects", req_body)
		res := httptest.NewRecorder()

		mock_project := &database.Project{
			Name: "Capybara Org",
		}

		jwt_validator.validate_return = user_id
		project_service.create_return = mock_project
		project_service.create_err = nil

		sid_cookie := &http.Cookie{
			Name: "sid",
			Value: user_id,
			Path: "/",
			SameSite: http.SameSiteStrictMode,
			MaxAge: 3600 * 24,
			HttpOnly: true,
			Secure: os.Getenv("STAGE") != "local",
		}
		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusCreated

		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
		
		var got_body utils.BaseResponse[database.Project]
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&got_body); err != nil {
			t.Errorf("got error %v, want error nil", err.Error())
		}

		got_project_name := got_body.Data.(map[string]any)["name"]
		want_org_name := project_name

		if got_project_name != want_org_name {
			t.Errorf("got project name %s, want %s", got_project_name, want_org_name)
		}

		got_called_n_times := project_service.create_n_calls
		want_called_n_times := 1
		got_called_with := project_service.create_call_args[0]
		want_called_with := []string{
			user_id,
			org_id,
			project_name,
		}

		if got_called_n_times != want_called_n_times {
			t.Errorf("got create called %d times, want %d", got_called_n_times, want_called_n_times)
		}

		for i, arg := range got_called_with {
			want_arg := want_called_with[i] 
			if arg != want_arg {
				t.Errorf("got arg %d %s, want %s", i + 1, arg, want_arg)
			}
		}
	})

	t.Run("it returns status code 400 when validation failed", func (t *testing.T) {
		tt := []struct{
			desc string
			body any
		}{
			{
				"empty json",
				"{}",
			},
			{
				"missing org_id",
				`
				{
					"name": "Shop with Sophia"
				}
				`,
			},
			{
				"missing name",
				`
				{
					"org_id": "885f535a-7a45-450b-8c1a-1cbd0f46e5d8"
				}
				`,
			},
			{
				"invalid org_id format",
				`
				{
					"org_id": "abcd",
					"name": "Study with Selma"
				}
				`,
			},
		}

		for _, tt := range tt {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				payload, _ := tt.body.(string)
				body := bytes.NewReader([]byte(payload))
				req, _ := http.NewRequest(http.MethodPost, "/api/projects", body)
				res := httptest.NewRecorder()

				sid_cookie := &http.Cookie{
					Name: "sid",
					Value: "123",
					Path: "/",
					SameSite: http.SameSiteStrictMode,
					MaxAge: 3600 * 24,
					HttpOnly: true,
					Secure: os.Getenv("STAGE") != "local",
				}
				req.AddCookie(sid_cookie)

				server.ServeHTTP(res, req)

				got_status := res.Result().StatusCode
				want_status := http.StatusBadRequest

				if got_status != want_status {
					t.Errorf("got status code %d, want %d", got_status, want_status)
				}
			})
		}
	})
}

func TestProjectGetOne(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
	project_service := &StubProjectService{}
	jwt_validator := &StubJwtValidator{}

	mock_user := &database.User{
		Email: "Salman",
		UserID: pgtype.UUID{},
		Username: "frnamlas",
		CreatedAt: pgtype.Timestamp{},
		UpdatedAt: pgtype.Timestamp{},
		FullName: "Salman RF",
	}
	user_service.find_by_id_return = mock_user
	user_service.find_by_id_err = nil

	server := api.NewAPIServer(
		test_ctx,
		user_service,
		auth_service,
		org_service,
		project_service,
		jwt_validator,
	)

	sid_cookie := &http.Cookie{
		Name: "sid",
		Value: "123",
		Path: "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge: 3600 * 24,
		HttpOnly: true,
		Secure: os.Getenv("STAGE") != "local",
	}
	
	project_service.find_by_id_return = &database.FindOneProjectByIdRow{
		Name: pgtype.Text{String: "Capybara"},
	}

	t.Run("it should return status 200 and the project", func (t *testing.T) {
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_project_uuid := pgtype.UUID{}
		mock_project_uuid.Scan(mock_project_id)

		project_service.find_by_id_return.ProjectID = mock_project_uuid
		
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%s", mock_project_id), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusOK

		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}

		var got_body map[string]interface{}
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&got_body); err != nil {
			t.Errorf("got error %s, want nil", err.Error())
		}

		got_id := got_body["data"].(map[string]interface{})["project_id"]
		want_id := mock_project_id

		if got_id != want_id {
			t.Errorf("got project_id %v, want %s", got_id, want_id)
		}
	})

	t.Run("it should return status 404 when project not found", func (t *testing.T) {
		mock_project_uuid := pgtype.UUID{}
		mock_project_uuid.Scan("28451bd5-0113-4ec6-9540-6646ae72a957")

		project_service.find_by_id_return = nil
		
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%s", mock_project_uuid.String()), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound

		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})
}

func TestProjectUpdateOne(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
	project_service := &StubProjectService{}
	jwt_validator := &StubJwtValidator{}

	mock_user := &database.User{
		Email: "Salman",
		UserID: pgtype.UUID{},
		Username: "frnamlas",
		CreatedAt: pgtype.Timestamp{},
		UpdatedAt: pgtype.Timestamp{},
		FullName: "Salman RF",
	}
	user_service.find_by_id_return = mock_user
	user_service.find_by_id_err = nil

	server := api.NewAPIServer(
		test_ctx,
		user_service,
		auth_service,
		org_service,
		project_service,
		jwt_validator,
	)

	sid_cookie := &http.Cookie{
		Name: "sid",
		Value: "123",
		Path: "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge: 3600 * 24,
		HttpOnly: true,
		Secure: os.Getenv("STAGE") != "local",
	}
	
	project_service.find_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
		Name: pgtype.Text{String: "Capybara", Valid: true},
	}

	t.Run("it should return status 200 and the updated project", func (t *testing.T) {
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_project_uuid := pgtype.UUID{}
		mock_project_uuid.Scan(mock_project_id)

		project_service.find_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
			ProjectID: mock_project_uuid,
			Name: pgtype.Text{String: "Capybara", Valid: true},
			Role: "owner",
		}

		new_name := "Binturong Org"

		project_service.update_one_return = &database.Project{
			ProjectID: mock_project_uuid,
			Name: new_name,
		}
		payload := fmt.Sprintf(`{"name": "%s"}`, new_name)
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%s", mock_project_id), body)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusOK
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
		
		var got_body utils.BaseResponse[any]
		decoder := json.NewDecoder(res.Result().Body)
		if err := decoder.Decode(&got_body); err != nil {
			t.Errorf("got error %s, want nil", err.Error())
		}
		
		got_name := got_body.Data.(map[string]any)["name"]
		want_id := new_name
		
		if got_name != want_id {
			t.Errorf("got name %v, want %s", got_name, want_id)
		}

		got_update_one_calls := project_service.update_one_n_calls
		want_update_one_n_calls := 1

		if got_update_one_calls != want_update_one_n_calls {
			t.Errorf("got update one called %v times, want %v", got_update_one_calls, want_update_one_n_calls)
		}
	})

	t.Run("it should return status 403 if doesn't have sufficient permission", func (t *testing.T) {
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_project_uuid := pgtype.UUID{}
		mock_project_uuid.Scan(mock_project_id)

		project_service.find_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
			ProjectID: mock_project_uuid,
			Name: pgtype.Text{String: "Capybara", Valid: true},
			Role: "member",
		}

		payload := `{"name": "Tai Lung"}`
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%s", mock_project_id), body)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusForbidden
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})

	t.Run("it should return status 404 if project id not provided", func (t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/api/projects/", nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound

		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})
	t.Run("it should return status 404 if project not found", func (t *testing.T) {
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_project_uuid := pgtype.UUID{}
		mock_project_uuid.Scan(mock_project_id)

		project_service.find_by_id_and_role_return = nil
		project_service.find_by_id_and_role_error = nil

		new_name := "Binturong Org"
		payload := fmt.Sprintf(`{"name": "%s"}`, new_name)
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%s", mock_project_id), body)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound

		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})
}

func TestProjectDeleteOne(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
	project_service := &StubProjectService{}
	jwt_validator := &StubJwtValidator{}

	mock_user := &database.User{
		Email: "Salman",
		UserID: pgtype.UUID{},
		Username: "frnamlas",
		CreatedAt: pgtype.Timestamp{},
		UpdatedAt: pgtype.Timestamp{},
		FullName: "Salman RF",
	}
	user_service.find_by_id_return = mock_user
	user_service.find_by_id_err = nil

	server := api.NewAPIServer(
		test_ctx,
		user_service,
		auth_service,
		org_service,
		project_service,
		jwt_validator,
	)

	sid_cookie := &http.Cookie{
		Name: "sid",
		Value: "123",
		Path: "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge: 3600 * 24,
		HttpOnly: true,
		Secure: os.Getenv("STAGE") != "local",
	}
	
	project_service.find_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
		Role: "owner",
	}

	t.Run("it should return status 204 on deletion", func (t *testing.T) {
		project_service.delete_one_n_calls = 0
		project_service.delete_one_call_args = []string{}
		
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_project_uuid := pgtype.UUID{}
		mock_project_uuid.Scan(mock_project_id)
		project_service.find_by_id_and_role_return.ProjectID = mock_project_uuid
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%s", mock_project_id), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNoContent
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}

		got_called := project_service.delete_one_n_calls
		want_called := 1
		got_called_with := project_service.delete_one_call_args[0]
		want_called_with := mock_project_id

		if got_called != want_called {
			t.Errorf("got update one called %d times, want %d", got_called, want_called)
		}

		if got_called_with != want_called_with {
			t.Errorf("got update one called with %s, want %s", got_called_with, want_called_with)
		}
	})

	t.Run("it should return status 404 when project_id is not specified", func (t *testing.T) {
		project_service.find_by_id_and_role_return = nil
		
		req, _ := http.NewRequest(http.MethodDelete, "/api/projects/", nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})

	t.Run("it should return status 204 when org not found (already deleted)", func (t *testing.T) {
		project_service.find_by_id_and_role_return = nil
		
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%s", mock_project_id), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNoContent
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})

	t.Run("it should return status 403 when doesn't have suficient permission", func (t *testing.T) {
		project_service.find_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
			Role: "owner",
		}
		
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%s", mock_project_id), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNoContent
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})

	t.Run("it should return status 403 when doesn't have suficient permission", func (t *testing.T) {
		project_service.find_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
			Role: "owner",
		}
		
		mock_project_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%s", mock_project_id), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNoContent
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})
}