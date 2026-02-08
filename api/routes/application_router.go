package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/api/middleware"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/pkg/auth"
)

func SetupApplicationRouter(application_service application.Service, jwt_validator auth.JWT) chi.Router {
	r := chi.NewRouter()
	
	app_handlers := handlers.NewAppHandlers(application_service)

	r.Post("/", middleware.LoginGuard(
		jwt_validator, 
		http.HandlerFunc(app_handlers.HandleCreate),
	))

	r.Get("/{app_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(app_handlers.HandleFindOne),
	))

	r.Put("/{app_id}", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(app_handlers.HandleUpdate),
	))

	r.Get("/{app_id}/configs", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(app_handlers.HandleFindOneConfig),
	))

	r.Post("/{app_id}/configs", middleware.LoginGuard(
		jwt_validator,
		http.HandlerFunc(app_handlers.HandleCreateConfig),
	))

	return r
}