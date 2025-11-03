package routes

import (
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/user"
)

func SetupAuthRouter(mux *http.ServeMux, as auth.Service, us user.Service) {
	mux.Handle("GET /api/auth/me", handlers.GetMeHandler(as, us))
	mux.Handle("POST /api/auth/signup", handlers.SignupHandler(as, us))
	mux.Handle("POST /api/auth/signin", handlers.SigninHandler(us))
}