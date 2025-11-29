package routes

import (
	"net/http"

	"github.com/salmanrf/capybara-cloud/api/handlers"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
)

func SetupAuthRouter(mux *http.ServeMux, as auth.Service, us user.Service, jwt_utils auth_utils.JWT) {
	mux.Handle("GET /api/auth/me", handlers.GetMeHandler(as, us, jwt_utils))
	mux.Handle("POST /api/auth/signup", handlers.SignupHandler(as, us))
	mux.Handle("POST /api/auth/signin", handlers.SigninHandler(us, jwt_utils))
}