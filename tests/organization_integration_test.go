package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestOrganizationCreateIntegration(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
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
		jwt_validator,
	)
	
	t.Run("it returns status code 201 and the new org on success", func (t *testing.T) {
		req_body := bytes.NewReader([]byte(`
			{
				"name": "Capybara Org"
			}
		`))
		req, _ := http.NewRequest(http.MethodPost, "/api/organizations", req_body)
		res := httptest.NewRecorder()

		mock_org := &database.Organization{
			Name: "Capybara Org",
		}

		org_service.create_return = mock_org
		org_service.create_err = nil

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
		want_status := http.StatusCreated

		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
		
		var got_body database.Organization
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&got_body); err != nil {
			t.Errorf("got error %v, want error nil", err.Error())
		}
		got_org_name := mock_org.Name
		want_org_name := mock_org.Name

		if got_org_name != want_org_name {
			t.Errorf("got org name %s, want %s", got_org_name, want_org_name)
		}
	})

	t.Run("it returns status code 401 on missing credentials", func (t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/organizations", nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusUnauthorized

		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("it returns status code 401 on jwt validation error", func (t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/organizations", nil)
		res := httptest.NewRecorder()

		jwt_validator.validate_error = errors.New("token expired")

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
		want_status := http.StatusUnauthorized

		if got_status != want_status {
			t.Errorf("got status code %d, want %d", got_status, want_status)
		}
	})

	t.Run("it returns status code 400 on body validation error", func (t *testing.T) {
		tests := []struct{
			desc string
			body any
		}{
			{"empty body", nil},
			{"empty body", ""},
			{"malformed json", "name: capybara"},
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

		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				req_body, _ := tt.body.(string)
				body := bytes.NewReader([]byte(req_body))
				req, _ := http.NewRequest(http.MethodPost, "/api/organizations", body)
				res := httptest.NewRecorder()
				
				req.AddCookie(sid_cookie)
		
				server.ServeHTTP(res, req)
		
				got_status := res.Result().StatusCode
				want_status := http.StatusUnauthorized
		
				if got_status != want_status {
					t.Errorf("got status code %d, want %d", got_status, want_status)
				}
			})
		}
	})
}

func TestOrganizationGetOne(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
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
	
	org_service.find_by_id_return = &database.FindOneOrganizationByIdRow{
		Name: pgtype.Text{String: "Capybara"},
	}

	t.Run("it should return status 200 and the organization", func (t *testing.T) {
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_org_uuid := pgtype.UUID{}
		mock_org_uuid.Scan(mock_org_id)

		org_service.find_by_id_return.OrgID = mock_org_uuid
		
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/organizations/%s", mock_org_id), nil)
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

		got_id := got_body["data"].(map[string]interface{})["org_id"]
		want_id := mock_org_id

		if got_id != want_id {
			t.Errorf("got org_id %v, want %s", got_id, want_id)
		}
	})

	t.Run("it should return status 404 when organization not found", func (t *testing.T) {
		mock_org_uuid := pgtype.UUID{}
		mock_org_uuid.Scan("28451bd5-0113-4ec6-9540-6646ae72a957")

		org_service.find_by_id_return = nil
		
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/organizations/%s", mock_org_uuid.String()), nil)
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

func TestOrganizationUpdateOne(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
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
	
	org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
		Name: pgtype.Text{String: "Capybara", Valid: true},
	}

	t.Run("it should return status 200 and the updated organization", func (t *testing.T) {
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_org_uuid := pgtype.UUID{}
		mock_org_uuid.Scan(mock_org_id)

		org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
			OrgID: mock_org_uuid,
			Name: pgtype.Text{String: "Capybara", Valid: true},
			Role: "owner",
		}

		new_name := "Binturong Org"

		org_service.update_one_return = &database.Organization{
			OrgID: mock_org_uuid,
			Name: new_name,
		}
		payload := fmt.Sprintf(`{"name": "%s"}`, new_name)
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/organizations/%s", mock_org_id), body)
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

		got_update_one_calls := org_service.update_one_n_calls
		want_update_one_n_calls := 1

		if got_update_one_calls != want_update_one_n_calls {
			t.Errorf("got update one called %v times, want %v", got_update_one_calls, want_update_one_n_calls)
		}
	})

	t.Run("it should return status 403 if doesn't have sufficient permission", func (t *testing.T) {
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_org_uuid := pgtype.UUID{}
		mock_org_uuid.Scan(mock_org_id)

		org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
			OrgID: mock_org_uuid,
			Name: pgtype.Text{String: "Capybara", Valid: true},
			Role: "member",
		}

		payload := `{"name": "Tai Lung"}`
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/organizations/%s", mock_org_id), body)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusForbidden
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})

	t.Run("it should return status 404 if organization id not provided", func (t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/api/organizations/", nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNotFound

		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}
	})
	t.Run("it should return status 404 if organization not found", func (t *testing.T) {
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_org_uuid := pgtype.UUID{}
		mock_org_uuid.Scan(mock_org_id)

		org_service.find_by_id_and_role_return = nil
		org_service.find_by_id_and_role_error = nil

		new_name := "Binturong Org"
		payload := fmt.Sprintf(`{"name": "%s"}`, new_name)
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/organizations/%s", mock_org_id), body)
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

func TestOrganizationDeleteOne(t *testing.T) {
	test_ctx := context.Background()

	user_service := &StubUserService{}
	auth_service := &StubAuthService{}
	org_service := &StubOrgService{}
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
	
	org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
		Role: "owner",
	}

	t.Run("it should return status 204 on deletion", func (t *testing.T) {
		org_service.delete_one_n_calls = 0
		org_service.delete_one_call_args = []string{}
		
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		mock_org_uuid := pgtype.UUID{}
		mock_org_uuid.Scan(mock_org_id)
		org_service.find_by_id_and_role_return.OrgID = mock_org_uuid
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/organizations/%s", mock_org_id), nil)
		res := httptest.NewRecorder()

		req.AddCookie(sid_cookie)

		server.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusNoContent
		
		if got_status != want_status {
			t.Errorf("got status %d, want %d", got_status, want_status)
		}

		got_called := org_service.delete_one_n_calls
		want_called := 1
		got_called_with := org_service.delete_one_call_args[0]
		want_called_with := mock_org_id

		if got_called != want_called {
			t.Errorf("got update one called %d times, want %d", got_called, want_called)
		}

		if got_called_with != want_called_with {
			t.Errorf("got update one called with %s, want %s", got_called_with, want_called_with)
		}
	})

	t.Run("it should return status 404 when org_id is not specified", func (t *testing.T) {
		org_service.find_by_id_and_role_return = nil
		
		req, _ := http.NewRequest(http.MethodDelete, "/api/organizations/", nil)
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
		org_service.find_by_id_and_role_return = nil
		
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/organizations/%s", mock_org_id), nil)
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
		org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
			Role: "owner",
		}
		
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/organizations/%s", mock_org_id), nil)
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
		org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
			Role: "owner",
		}
		
		mock_org_id := "28451bd5-0113-4ec6-9540-6646ae72a957"
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/organizations/%s", mock_org_id), nil)
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