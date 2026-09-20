package handlers

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"os"
	"time"

	auth_module "github.com/salmanrf/capybara-cloud/apps/backend/internal/auth"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/apps/backend/pkg/auth"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"

	pkgerr "github.com/pkg/errors"
)

type auth_handler struct {
	logger *slog.Logger
	auth_service auth_module.Service
	user_service user.Service
	jwt_utils    utils.JWT
}

type AuthHandlers interface {
	HandleGetMe(w http.ResponseWriter, r *http.Request)
	HandleSignup(w http.ResponseWriter, r *http.Request)
	HandleSignin(w http.ResponseWriter, r *http.Request)
}

func NewAuthHandlers(logger *slog.Logger, auth_service auth_module.Service, user_service user.Service, jwt_utils utils.JWT) AuthHandlers {
	return &auth_handler{
		logger,
		auth_service,
		user_service,
		jwt_utils,
	}
}

func (h *auth_handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	sid_cookie, err := r.Cookie("sid")

	if err != nil {
		h.logger.Error("[HandleGetMe] Missing session cookie", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	sub, err := h.jwt_utils.ValidateJWT(sid_cookie.Value, os.Getenv("AUTH_JWT_SECRET"))

	if err != nil {
		h.logger.Error("[HandleGetMe] Invalid session JWT", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	user, err := h.auth_service.GetMe(sub)

	if err != nil {
		h.logger.Error("[HandleGetMe] Unable to find user", "error", pkgerr.WithStack(err), "user_id", sub)
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
		h.logger.Error("[HandleSignup] Unable to parse request", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
		return
	}

	_, err := body.Validate()
	if err != nil {
		h.logger.Error("[HandleSignup] Unable to validate request", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	existing, err := h.user_service.FindById(body.Email, true)

	if err != nil {
		h.logger.Error("[HandleSignup] Unable to find existing user", "error", pkgerr.WithStack(err), "email", body.Email)
		utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
		return
	}

	if existing != nil {
		h.logger.Error("[HandleSignup] User already exists", "email", body.Email)
		utils.ResponseWithError(w, http.StatusBadRequest, nil, "This user already exists")
		return
	}

	_, err = h.user_service.Create(body)

	if err != nil {
		h.logger.Error("[HandleSignup] Unable to create user", "error", pkgerr.WithStack(err), "email", body.Email, "username", body.Username)
		utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
		return
	}

	utils.ResponseWithSuccess[any](w, http.StatusOK, nil, "Signed up successfully")
}

func (h *auth_handler) HandleSignin(w http.ResponseWriter, r *http.Request) {
	var body dto.SigninDto

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		h.logger.Error("[HandleSignin] Unable to parse request", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
		return
	}

	_, err := body.Validate()
	if err != nil {
		h.logger.Error("[HandleSignin] Unable to validate request", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	user, err := h.user_service.FindById(body.Email, true)

	if err != nil {
		h.logger.Error("[HandleSignin] Unable to find user", "error", pkgerr.WithStack(err), "email", body.Email)
		utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
		return
	}

	if user == nil {
		h.logger.Error("[HandleSignin] User not found", "email", body.Email)
		utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
		return
	}

	password_match, err := auth_utils.HashCompare(body.Password, user.HashedPassword)

	if err != nil {
		h.logger.Error("[HandleSignin] Unable to compare password hash", "error", pkgerr.WithStack(err), "user_id", user.UserID)
		utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
		return
	}

	if !password_match {
		h.logger.Error("[HandleSignin] Password mismatch", "user_id", user.UserID)
		utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
		return
	}

	jwt_string, err := h.jwt_utils.MakeJWT(user.UserID.String(), os.Getenv("AUTH_JWT_SECRET"), time.Hour * 24)

	if err != nil {
		h.logger.Error("[HandleSignin] Unable to build JWT", "error", pkgerr.WithStack(err), "user_id", user.UserID)
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