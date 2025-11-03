package routes

import (
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/organization"
)

func SetupOrganizationRouter(mux *http.ServeMux, os organization.Service) {
	mux.Handle(
		"POST /api/organizations", 
		middleware.LoginGuard(handlers.CreateOrgHandler(os)),
	)
	mux.Handle(
		"GET /api/organizations", 
		middleware.LoginGuard(handlers.ListMyOrgHandler(os)),
	)
}