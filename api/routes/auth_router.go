package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/salmanrf/capybara-cloud/api/handlers"
	auth_module "github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/user"
	"github.com/salmanrf/capybara-cloud/shared/utils"
)

func SetupAuthRouter(auth_service auth_module.Service, user_service user.Service, jwt_utils utils.JWT) chi.Router {
	r := chi.NewRouter()

	auth_handlers := handlers.NewAuthHandlers(auth_service, user_service, jwt_utils)

	r.Get("/me", http.HandlerFunc(auth_handlers.HandleGetMe))
	r.Post("/signup", http.HandlerFunc(auth_handlers.HandleSignup))
	r.Post("/signin", http.HandlerFunc(auth_handlers.HandleSignin))

	return r
}