package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/pkg/auth"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func SetupApplicationRouter(mux *http.ServeMux, application_service application.Service, jwt_validator auth.JWT) {
	mux.Handle(
		"POST /api/applications", 
		middleware.LoginGuard(
			jwt_validator, 
			handlers.CreateApplicationHandler(application_service),
		),
	)

	find_one_handler := handlers.FindOneApplicationHandler(application_service)

	mux.Handle(
		"GET /api/applications/",
		middleware.LoginGuard(
			jwt_validator,
			http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
				app_id := strings.TrimPrefix(r.URL.Path, "/api/applications/")

				if app_id != "" {
					ctx := context.WithValue(r.Context(), "app_id", app_id)
					new_req := r.WithContext(ctx)
					find_one_handler.ServeHTTP(w, new_req)
					return
				}

				utils.ResponseWithError(
					w,
					http.StatusNotFound,
					nil,
					"Not found",
				)
			}),
		),
	)

	mux.Handle(
		"PUT /api/applications/",
		middleware.LoginGuard(
			jwt_validator,
			handlers.UpdateApplicationHandler(application_service),
		),
	)
}