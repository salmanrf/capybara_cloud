package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/api/routes"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
)

type api_server struct {
	http.Handler
}

func NewAPIServer(
	ctx context.Context,
	application_service application.Service,
	user_service user.Service,
	auth_service auth.Service,
	org_service organization.Service,
	project_service project.Service,
	jwt_validator auth_utils.JWT,
) http.Handler {
	router := chi.NewRouter()

	router.Route("/api", func (r chi.Router) {
		r.Mount("/applications", routes.SetupApplicationRouter(
			application_service,
			jwt_validator,
		))
		r.Mount("/organizations", routes.SetupOrganizationRouter(
			org_service,
			jwt_validator,
		))
		r.Mount("/auth", routes.SetupAuthRouter(
			auth_service,
			user_service,
			jwt_validator,
		))
		r.Mount("/projects", routes.SetupProjectRouter(
			project_service,
			jwt_validator,
		))
	})

	s := api_server{
		router,
	}

	return s
}