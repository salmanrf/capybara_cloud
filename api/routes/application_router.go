package routes

import (
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/pkg/auth"
)

func SetupApplicationRouter(mux *http.ServeMux, application_service application.Service, jwt_validator auth.JWT) {
	mux.Handle(
		"POST /api/applications", 
		middleware.LoginGuard(
			jwt_validator, 
			handlers.CreateApplicationHandler(application_service),
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