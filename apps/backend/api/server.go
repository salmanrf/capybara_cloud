package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/apps/backend/api/middleware"
	"github.com/salmanrf/capybara-cloud/apps/backend/api/routes"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/application"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/auth"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/deployment"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/organization"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/project"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/user"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

type api_server struct {
	http.Handler
}

func NewAPIServer(
	logger *slog.Logger,
	application_service application.Service,
	deployment_service deployment.Service,
	user_service user.Service,
	auth_service auth.Service,
	org_service organization.Service,
	project_service project.Service,
	jwt_validator utils.JWT,
) http.Handler {
	router := chi.NewRouter()

	loggingmd := middleware.CreateLoggingMiddleware(logger)

	router.Use(loggingmd)

	router.Route("/api", func (r chi.Router) {
		r.Mount("/applications", routes.SetupApplicationRouter(
			logger,
			application_service,
			deployment_service,
			jwt_validator,
		))
		r.Mount("/organizations", routes.SetupOrganizationRouter(
			logger,
			org_service,
			jwt_validator,
		))
		r.Mount("/auth", routes.SetupAuthRouter(
			logger,
			auth_service,
			user_service,
			jwt_validator,
		))
		r.Mount("/projects", routes.SetupProjectRouter(
			logger,
			project_service,
			jwt_validator,
		))
	})

	s := api_server{
		router,
	}

	return s
}