package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/apps/backend/api/handlers"
	"github.com/salmanrf/capybara-cloud/apps/backend/api/middleware"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/organization"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

func SetupOrganizationRouter(org_service organization.Service, jwt_validator utils.JWT) chi.Router {
	r := chi.NewRouter()

	org_handlers := handlers.NewOrgHandlers(org_service)

	// Handle base path - GET for list, POST for create
	r.Post("/", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(org_handlers.HandleCreate),
	))

	r.Get("/", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(org_handlers.HandleListMyOrganizations),
	))

	// Handle requests to base path with methods that require org_id
	r.Put("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseWithError(w, http.StatusNotFound, nil, "Organization ID required")
	}))

	r.Delete("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseWithError(w, http.StatusNotFound, nil, "Organization ID required")
	}))

	// Routes for specific organization
	r.Get("/{org_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(org_handlers.HandleFindOne),
	))

	r.Put("/{org_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(org_handlers.HandleUpdate),
	))

	r.Delete("/{org_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(org_handlers.HandleDelete),
	))

	return r
}