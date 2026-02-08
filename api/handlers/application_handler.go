package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

type app_handler struct {
	app_service application.Service
}

type AppHandlers interface {
	HandleFindOne(w http.ResponseWriter, r *http.Request) 
	HandleCreate(w http.ResponseWriter, r *http.Request)
	HandleUpdate(w http.ResponseWriter, r *http.Request)
	HandleCreateConfig(w http.ResponseWriter, r *http.Request)
}

func NewAppHandlers(app_service application.Service) AppHandlers {
	return &app_handler{
		app_service,
	}
}

func (h *app_handler) HandleFindOne(w http.ResponseWriter, r *http.Request) {
	app_id := r.PathValue("app_id")
	user_id, _ := r.Context().Value("user_id").(string)

	app, err :=  h.app_service.FindOne(app_id, user_id)

	if err != nil {
		errmsg := err.Error()
		switch errmsg {
		case "permission_denied":
			utils.ResponseWithError(
				w,
				http.StatusForbidden,
				nil,
				"Internal server error",
			)
			return
		case "not_found":
			utils.ResponseWithError(
				w,
				http.StatusInternalServerError,
				nil,
				"Internal server error",
			)
			return
		default:
			utils.ResponseWithError(
				w,
				http.StatusInternalServerError,
				nil,
				"Internal server error",
			)
			return	
		}
	}
	
	if app == nil {
		utils.ResponseWithError(
			w,
			http.StatusNotFound,
			nil,
			"Not found",
		)
		return
	}
	
	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		&app,
		"Application retrieved successfully",
	)
}

func (h *app_handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var body dto.CreateApplicationDto

	rctx := r.Context()
	user_id, _ := rctx.Value("user_id").(string)

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			"unprocessable entity",
		)
		return
	}

	_, err := body.Validate()
	if err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	
	new_application, err := h.app_service.Create(
		user_id,
		body,
	)

	if err != nil {
		if err.Error() == "permission_denied" {
			utils.ResponseWithError(
				w, 
				http.StatusForbidden, 
				nil,
				"insufficient permission to create app on this project", 
			)
		} else {
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "internal server error")
		}
		return
	}

	utils.ResponseWithSuccess(
		w,
		http.StatusCreated,
		new_application,
		"Application created successfully",
	)
}

func (h *app_handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	app_id := r.PathValue("app_id")

	user_id, _ := r.Context().Value("user_id").(string)
	if user_id == "" {
		utils.ResponseWithError(
			w,
			http.StatusUnauthorized,
			nil,
			"Unauthorized",
		)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var body dto.UpdateApplicationDto
	if err := decoder.Decode(&body); err != nil {
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			err.Error(),
		)
		return	
	}
	if _, err := body.Validate(); err != nil {
		utils.ResponseWithError(
			w,
			http.StatusBadRequest,
			nil,
			err.Error(),
		)
		return	
	}

	updated_app, err := h.app_service.Update(
		app_id,
		user_id,
		body,
	)

	if err != nil {
		errmsg := err.Error()
		if errmsg == "permission_denied" {
			utils.ResponseWithError(
				w,
				http.StatusForbidden,
				nil,
				"Insufficient permission to update application",
			)
			return	
		} 
		if errmsg == "not_found" {
			utils.ResponseWithError(
				w,
				http.StatusNotFound,
				nil,
				"Not found",
			)
			return
		}
		utils.ResponseWithError(
				w,
				http.StatusInternalServerError,
				nil,
				"Internal server error",
			)
			return
	}
	
	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		&updated_app,
		"Application updated successfully",
	)
}

func (h *app_handler) HandleCreateConfig(w http.ResponseWriter, r *http.Request) {
	app_id := r.PathValue("app_id")
	user_id := r.Context().Value("user_id").(string)
	
	var body dto.CreateApplicationConfigDto
	
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			"unprocessable entity",
		)
		return
	}

	_, err := body.Validate()
	if err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	app_cfg, err := h.app_service.CreateConfig(
		app_id,
		user_id,
		body,
	)
	if err != nil {
		utils.ResponseWithError(
			w,
			http.StatusInternalServerError,
			nil,
			err.Error(),
		)
		return
	}
	
	app_config_response := dto.ApplicationConfigResponse{
		AppCfgID: app_cfg.AppCfgID.String(),
		AppID: app_cfg.AppID.String(),
		VariablesJson: string(app_cfg.VariablesJson),
		ConfigVariables: body.Variables,
		CreatedAt: app_cfg.CreatedAt.Time,
		UpdatedAt: app_cfg.UpdatedAt.Time,
	}

	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		&app_config_response,
		"Bad request",
	)
}