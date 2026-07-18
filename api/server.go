package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/api/routes"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/deployment"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/internal/user"
	"github.com/salmanrf/capybara-cloud/shared/utils"
)

type api_server struct {
	http.Handler
}

func NewAPIServer(
	ctx context.Context,
	application_service application.Service,
	deployment_service deployment.Service,
	user_service user.Service,
	auth_service auth.Service,
	org_service organization.Service,
	project_service project.Service,
	jwt_validator utils.JWT,
) http.Handler {
	router := chi.NewRouter()

	router.Route("/api", func (r chi.Router) {
		r.Mount("/applications", routes.SetupApplicationRouter(
			application_service,
			deployment_service,
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