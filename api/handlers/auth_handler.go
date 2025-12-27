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

type auth_handler struct {
	auth_service auth_module.Service
	user_service user.Service
	jwt_utils    auth_utils.JWT
}

type AuthHandlers interface {
	HandleGetMe(w http.ResponseWriter, r *http.Request)
	HandleSignup(w http.ResponseWriter, r *http.Request)
	HandleSignin(w http.ResponseWriter, r *http.Request)
}

func NewAuthHandlers(auth_service auth_module.Service, user_service user.Service, jwt_utils auth_utils.JWT) AuthHandlers {
	return &auth_handler{
		auth_service,
		user_service,
		jwt_utils,
	}
}

func (h *auth_handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	sid_cookie, err := r.Cookie("sid")

	if err != nil {
		fmt.Println("Error extracting session cookie", err.Error())
		utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	sub, err := h.jwt_utils.ValidateJWT(sid_cookie.Value, os.Getenv("AUTH_JWT_SECRET"))

	if err != nil {
		fmt.Println("Error validating session JWT", err.Error())
		utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	user, err := h.auth_service.GetMe(sub)

	if err != nil {
		utils.ResponseWithError(
			w,
			http.StatusNotFound,
			nil,
			"User not found",
		)
		return
	}

	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		dto.NewAuthMeResponse(user),
		"Session retrieved successfully",
	)
}

func (h *auth_handler) HandleSignup(w http.ResponseWriter, r *http.Request) {
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

	existing, err := h.user_service.FindById(body.Email, true)

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

	_, err = h.user_service.Create(body)

	if err != nil {
		fmt.Println("Signup failed", err.Error())
		utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
	}

	utils.ResponseWithSuccess[any](w, http.StatusOK, nil, "Signed up successfully")
}

func (h *auth_handler) HandleSignin(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.user_service.FindById(body.Email, true)

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

	jwt_string, err := h.jwt_utils.MakeJWT(user.UserID, os.Getenv("AUTH_JWT_SECRET"), time.Hour * 24)

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