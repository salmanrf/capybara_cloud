package routes

import (
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/internal/organization"
)

func SetupOrganizationRouter(mux *http.ServeMux, s organization.Service) {
	mux.Handle("POST /api/organizations/", handlers.CreateOrgHandler(s))
}