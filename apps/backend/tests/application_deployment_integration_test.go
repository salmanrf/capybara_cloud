package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/apps/backend/api/routes"
	config "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

func TestCreateApplicationDeployment(t *testing.T) {
	app_service := &StubApplicationService{}
	deployment_service := &StubDeploymentService{}
	jwt_validator := &StubJwtValidator{}
	
	mux := chi.NewRouter()
	mux.Mount("/api/applications", 
		routes.SetupApplicationRouter(
			app_service, 
			deployment_service,
			jwt_validator,
		),
	)

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

	t.Run("should return status 401 if not logged in", func (t *testing.T) {
		defer func () {
			deployment_service.Clear()
		}()
		
		expected_app_id := "b87fcac7-05bc-4342-ad43-96c6e3c8afa3"
		req, _ := http.NewRequest(
			"POST", 
			fmt.Sprintf("/api/applications/%s/deployments", expected_app_id),
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

	t.Run("should return status 400/422 when validation failed", func (t *testing.T) {
		conf, _ := config.LoadConfig("./.env.test")
		conf.MAX_DEPLOY_BUNDLE_SIZE = 100
		conf.MAX_DEPLOY_FORM_SIZE = 100
		config.SetConfig(conf)
		
		expected_app_id := "b87fcac7-05bc-4342-ad43-96c6e3c8afa3"

		tests := []struct{
			desc string
			headers map[string]string
			datamap map[string]string
		}{
			{
				"invalid content-type",
				map[string]string{
					"Content-Type": "application/json",
				},
				map[string]string{},
			},
			{
				"empty body",
				map[string]string{
					"Content-Type": "multipart/form-data",
				},
				map[string]string{},
			},
			{
				"missing deployment bundle",
				map[string]string{
					"Content-Type": "multipart/form-data;", 
				},
				map[string]string{
					"bundle": "...",
				},
			},
		}
		
		for _, tt := range tests {
			t.Run(fmt.Sprintf("returns 400 on %s", tt.desc), func (t *testing.T) {
				defer func () {
					deployment_service.Clear()
				}()
				
				formdata := utils.NewMultipartForm()
				for k, v := range tt.datamap {
					formdata.SetField(k, v)
				}
				body := bytes.NewBuffer(formdata.GetEncoded())

				req, _ := http.NewRequest(
					"POST", 
					fmt.Sprintf("/api/applications/%s/deployments", expected_app_id),
					body,
				)
				for k, v := range tt.headers {
					req.Header.Add(k, v)
				}
				req.Header["Content-Type"] = []string{formdata.GetHeaderContentType()}
				req.Header.Add("Content-Length", formdata.GetHeaderContentLength())
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

	t.Run("should return status 413 on formdata/bundle file size too large", func (t *testing.T) {
		expected_app_id := "b87fcac7-05bc-4342-ad43-96c6e3c8afa3"

		tests := []struct{
			max_form_size int
			max_bundle_size int
			formsize int
			bundlesize int
		}{
			{
				100,
				90,
				101,
				92,
			},
			{
				100,
				50,
				1000,
				20,
			},
			{
				50,
				5,
				100,
				10,
			},
		}
		
		for _, tt := range tests {
			t.Run("returns 413 on data too large", func (t *testing.T) {
				defer func () {
					deployment_service.Clear()
				}()
				
				cfg, _ := config.LoadConfig("./.env.test")
				cfg.MAX_DEPLOY_FORM_SIZE = tt.max_form_size
				cfg.MAX_DEPLOY_BUNDLE_SIZE = tt.max_bundle_size
				config.SetConfig(cfg)

				formdata := utils.NewMultipartForm()
				formdata.SetFile("bundle", "bundle.tar.gz", tt.bundlesize, "application/gzip")
				body := bytes.NewBuffer(formdata.GetEncoded())
				req, _ := http.NewRequest(
					"POST", 
					fmt.Sprintf("/api/applications/%s/deployments", expected_app_id),
					body,
				)
				req.Header["Content-Type"] = []string{formdata.GetHeaderContentType()}
				req.Header.Add("Content-Length", fmt.Sprintf("%d", tt.formsize))
				req.AddCookie(sid_cookie)

				res := httptest.NewRecorder()
				api.ServeHTTP(res, req)
		
				got_status := res.Result().StatusCode
				want_status := http.StatusRequestEntityTooLarge
		
				if got_status != want_status {
					t.Errorf("got status code %d, want %v\n", got_status, want_status)
				}
			})
		}
	})

	t.Run("should return status 200 on successful deployment", func (t *testing.T) {
		conf, _ := config.LoadConfig("./.env.test")
		conf.MAX_DEPLOY_BUNDLE_SIZE = 500
		conf.MAX_DEPLOY_FORM_SIZE = 500
		config.SetConfig(conf)
		
		defer func () {
			deployment_service.Clear()
		}()
		
		config.LoadConfig("./.env.test")
		expected_app_id := "b87fcac7-05bc-4342-ad43-96c6e3c8afa3"
		formdata := utils.NewMultipartForm()
		formdata.SetFile("bundle", "bundle.tar.gz", 100, "application/gzip")
		body := bytes.NewBuffer(formdata.GetEncoded())

		req, _ := http.NewRequest(
			"POST", 
			fmt.Sprintf("/api/applications/%s/deployments", expected_app_id),
			body,
		)
		req.Header["Content-Type"] = []string{formdata.GetHeaderContentType()}
		req.Header.Add("Content-Length", formdata.GetHeaderContentLength())
		req.AddCookie(sid_cookie)
		
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)

		got_status := res.Result().StatusCode
		want_status := http.StatusCreated

		got_deploy_called_n := deployment_service.deploy_n_calls
		want_deploy_called_n := 1
		got_deploy_err := deployment_service.deploy_err
		var want_deploy_err error = nil

		if got_status != want_status {
			t.Errorf("got status code %d, want %d\n", got_status, want_status)
		}

		if got_deploy_called_n != want_deploy_called_n {
			t.Errorf("got deploy called %d times, want %d\n", got_deploy_called_n, want_deploy_called_n)
		}

		if got_deploy_err != want_deploy_err {
			t.Errorf("got deploy error %v, want %v\n", got_deploy_err, want_deploy_err)
		}
	})
}