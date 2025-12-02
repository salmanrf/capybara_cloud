package routes

import (
	"net/http"
	"strings"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/project"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
)

func SetupProjectRouter(mux *http.ServeMux, ps project.Service, jwt_validator auth_utils.JWT) {
	mux.Handle(
		"POST /api/projects", 
		middleware.LoginGuard(jwt_validator, handlers.CreateProjectHandler(ps)),
	)

	get_one_handler := handlers.GetOneProjectHandler(ps)
	list_handler := handlers.ListMyProjectHandler(ps)
	
	mux.Handle(
		"GET /api/projects/", 
		middleware.LoginGuard(jwt_validator, http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			endpoint := strings.TrimPrefix(r.URL.Path, "/api/projects") 

			if endpoint == "/" {
				list_handler(w, r)
				return
			}

			get_one_handler(w, r)
		})),
	)

	mux.Handle(
		"PUT /api/projects/", 
		middleware.LoginGuard(jwt_validator, handlers.UpdateOneProjectHandler(ps)),
	)

	mux.Handle(
		"DELETE /api/projects/", 
		middleware.LoginGuard(jwt_validator, handlers.DeleteOneProjectHandler(ps)),
	)
}