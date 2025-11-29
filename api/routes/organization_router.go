package routes

import (
	"net/http"
	"strings"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
)

func SetupOrganizationRouter(mux *http.ServeMux, os organization.Service, jwt_validator auth_utils.JWT) {
	mux.Handle(
		"POST /api/organizations", 
		middleware.LoginGuard(jwt_validator, handlers.CreateOrgHandler(os)),
	)

	get_one_handler := handlers.GetOneOrgHandler(os)
	list_handler := handlers.ListMyOrgHandler(os)
	
	mux.Handle(
		"GET /api/organizations/", 
		middleware.LoginGuard(jwt_validator, http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			endpoint := strings.TrimPrefix(r.URL.Path, "/api/organizations") 

			if endpoint == "/" {
				list_handler(w, r)
				return
			}

			get_one_handler(w, r)
		})),
	)

	mux.Handle(
		"PUT /api/organizations/", 
		middleware.LoginGuard(jwt_validator, handlers.UpdateOneOrgHandler(os)),
	)

	mux.Handle(
		"DELETE /api/organizations/", 
		middleware.LoginGuard(jwt_validator, handlers.DeleteOneOrgHandler(os)),
	)
}