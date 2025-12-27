package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/pkg/auth"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func SetupProjectRouter(project_service project.Service, jwt_validator auth.JWT) chi.Router {
	r := chi.NewRouter()

	project_handlers := handlers.NewProjectHandlers(project_service)

	r.Put("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseWithError(w, http.StatusNotFound, nil, "Project ID required")
	}))

	r.Delete("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseWithError(w, http.StatusNotFound, nil, "Project ID required")
	}))

	r.Get("/", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(project_handlers.HandleListMyProjects),
	))

	r.Post("/", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(project_handlers.HandleCreate),
	))

	r.Get("/{project_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(project_handlers.HandleFindOne),
	))

	r.Put("/{project_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(project_handlers.HandleUpdate),
	))

	r.Delete("/{project_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(project_handlers.HandleDelete),
	))

	return r
}