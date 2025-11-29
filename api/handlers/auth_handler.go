package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	auth_module "github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func GetMeHandler(as auth_module.Service, us user.Service, jwt_utils auth_utils.JWT) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid_cookie, err := r.Cookie("sid")

		if err != nil {
			fmt.Println("Error extracting session cookie", err.Error())
			utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}

		sub, err := jwt_utils.ValidateJWT(sid_cookie.Value, os.Getenv("AUTH_JWT_SECRET"))

		if err != nil {
			fmt.Println("Error validating session JWT", err.Error())	
			utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}

		user, err := as.GetMe(sub)
		
		if err != nil {
			utils.ResponseWithError(
				w,
				http.StatusNotFound,
				nil,
				"User not found",
			)
			return
		}

		fmt.Println("USER", user)

		utils.ResponseWithSuccess(
			w, 
			http.StatusOK, 
			dto.NewAuthMeResponse(user), 
			"Session retrieved successfully",
		)
	}
}

func SignupHandler(as auth_module.Service, us user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body dto.SignupDto

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&body); err != nil {
			fmt.Println("Signup failed, parsing", err.Error())
			utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
			return
		}

		_, err := body.Validate()
		if err != nil {
			fmt.Println("Signup failed, validation", err.Error())
			utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		existing, err := us.FindById(body.Email, true)

		if err != nil {
			fmt.Println("Singup failed", err.Error())
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
			return
		}

		if existing != nil {
			fmt.Println("Signup failed, user already exists")
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "This user already exists")
			return
		}

		_, err = us.Create(body)

		if err != nil {
			fmt.Println("Signup failed", err.Error())
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
		}

		utils.ResponseWithSuccess[any](w, http.StatusOK, nil, "Signed up successfully")
	}
}

func SigninHandler(us user.Service, jwt_utils auth_utils.JWT) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body dto.SigninDto 

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&body); err != nil {
			fmt.Println("Failed to decode request body", err)

			utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
			return
		}

		_, err := body.Validate()
		if err != nil {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		user, err := us.FindById(body.Email, true)

		if err != nil {
			fmt.Println("Error finding user", err)
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		if user == nil {
			fmt.Println("Error finding user")
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		password_match, err := auth_utils.HashCompare(body.Password, user.HashedPassword)

		if err != nil {
			fmt.Println("Error password compare", err)
			
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		if !password_match {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		jwt_string, err := jwt_utils.MakeJWT(user.UserID, os.Getenv("AUTH_JWT_SECRET"), time.Hour * 24) 

		if err != nil {
			fmt.Println("Error building jwt for signin", err.Error())
			
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
			return
		}

		sid_cookie := http.Cookie{
			Name: "sid",
			Value: jwt_string,
			Path: "/",
			SameSite: http.SameSiteStrictMode,
			MaxAge: 3600 * 24,
			HttpOnly: true,
			Secure: os.Getenv("STAGE") != "local",
		}

		http.SetCookie(w, &sid_cookie)

		utils.ResponseWithSuccess[any](
			w, 
			http.StatusOK, 
			nil,
			"Signed in successfully",
		)
	}
}