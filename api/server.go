package api

import (
	"context"
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/routes"
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
	user_service user.Service,
	auth_service auth.Service,
	org_service organization.Service,
	project_service project.Service,
	jwt_validator auth_utils.JWT,
) http.Handler {
	mux := http.NewServeMux()

	routes.SetupAuthRouter(mux, auth_service, user_service, jwt_validator)
	routes.SetupOrganizationRouter(mux, org_service, jwt_validator)
	routes.SetupProjectRouter(mux, project_service, jwt_validator)
	
	s := api_server{
		mux,
	}

	return s
}