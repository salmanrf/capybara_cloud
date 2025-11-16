package api

import (
	"context"
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/routes"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/internal/user"
)

type api_server struct {
	http.Handler
}

func NewAPIServer(
	ctx context.Context,
	user_service user.Service,
	auth_service auth.Service,
	org_service organization.Service,
) http.Handler {
	mux := http.NewServeMux()

	routes.SetupAuthRouter(mux, auth_service, user_service)
	routes.SetupOrganizationRouter(mux, org_service)
	
	s := api_server{
		mux,
	}

	return s
}