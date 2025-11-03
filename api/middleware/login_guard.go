package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/salmanrf/capybara-cloud/pkg/auth"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func LoginGuard(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid_cookie, err := r.Cookie("sid")
		if err != nil {
			fmt.Println("LoginGuard check failed", err.Error())
			utils.ResponseWithError(
				w,
				http.StatusUnauthorized,
				nil,
				"Unauthorized",
			)
			return 
		}

		sub, err := auth.ValidateJWT(sid_cookie.Value, os.Getenv("AUTH_JWT_SECRET"))
		if err != nil {
			fmt.Println("LoginGuard check failed", err.Error())
			utils.ResponseWithError(
				w,
				http.StatusUnauthorized,
				nil,
				"Unauthorized",
			)
			return 
		}

		new_ctx := context.WithValue(r.Context(), "user_id", sub)
		new_request := r.WithContext(new_ctx)
		
		next.ServeHTTP(w, new_request)
	}
}