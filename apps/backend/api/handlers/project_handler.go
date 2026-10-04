package handlers

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/internal/project"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"

	pkgerr "github.com/pkg/errors"
)

type project_handler struct {
	logger *slog.Logger
	project_service project.Service
}

type ProjectHandlers interface {
	HandleCreate(w http.ResponseWriter, r *http.Request)
	HandleFindOne(w http.ResponseWriter, r *http.Request)
	HandleListMyProjects(w http.ResponseWriter, r *http.Request)
	HandleUpdate(w http.ResponseWriter, r *http.Request)
	HandleDelete(w http.ResponseWriter, r *http.Request)
}

func NewProjectHandlers(logger *slog.Logger, project_service project.Service) ProjectHandlers {
	return &project_handler{
		logger,
		project_service,
	}
}

func (h *project_handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var body dto.CreateProjectDto

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		h.logger.Error("[HandleCreate] Unable to parse request", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			"Unprocessable Entity",
		)
		return
	}

	_, err := body.Validate()
	if err != nil {
		h.logger.Error("[HandleCreate] Unable to validate request", "error", pkgerr.WithStack(err))
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	project, err := h.project_service.Create(user_id, body.OrgId, body.Name)
	if err != nil {
		h.logger.Error("[HandleCreate] Unable to create project", "error", pkgerr.WithStack(err), "user_id", user_id, "org_id", body.OrgId, "project_name", body.Name)
		errmsg := err.Error()
		switch {
		case errmsg == "invalid_role":
			utils.ResponseWithError(w, http.StatusForbidden, nil, "Insufficient permission to create project")
		case strings.Contains(errmsg, "duplicate key"):
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Project with this name already exists")
		default:
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
		}
		return
	}

	utils.ResponseWithSuccess(
		w,
		http.StatusCreated,
		project,
		"Project created successfully",
	)
}

func (h *project_handler) HandleFindOne(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	project_id := r.PathValue("project_id")
	if project_id == "" {
		utils.ResponseWithSuccess[any](
			w,
			http.StatusNotFound,
			nil,
			"Project id not specified",
		)
		return
	}

	project, err := h.project_service.FindById(user_id, project_id)

	if err != nil {
		h.logger.Error("[HandleFindOne] Unable to find project", "error", pkgerr.WithStack(err), "user_id", user_id, "project_id", project_id)
		utils.ResponseWithSuccess[any](
			w,
			http.StatusOK,
			nil,
			err.Error(),
		)
		return
	}

	if project == nil {
		utils.ResponseWithError(
			w,
			http.StatusNotFound,
			nil,
			"Project not found",
		)
		return
	}

	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		project,
		"Project retrieved successfully",
	)
}

func (h *project_handler) HandleListMyProjects(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	org_id := r.URL.Query().Get("org_id")
	if org_id != "" {
		org_uuid := pgtype.UUID{}
		if err := org_uuid.Scan(org_id); err != nil {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "invalid org_id format, must be a valid uuid string")
			return
		}
	}

	projectuses, err := h.project_service.ListMyProjects(user_id, org_id)
	if err != nil {
		h.logger.Error("[HandleListMyProjects] Unable to list projects", "error", pkgerr.WithStack(err), "user_id", user_id)
		utils.ResponseWithError(
			w,
			http.StatusInternalServerError,
			nil,
			"Internal server error",
		)
		return
	}

	formatted := dto.NewListMyProjectResponse(projectuses)

	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		&formatted,
		"Project users retrieved successfuly",
	)
}

func (h *project_handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	project_id := r.PathValue("project_id")
	if project_id == "" {
		utils.ResponseWithSuccess[any](
			w,
			http.StatusNotFound,
			nil,
			"Project id not specified",
		)
		return
	}

	var body dto.UpdateProjectDto

	if r.Body == nil {
		h.logger.Error("[HandleUpdate] Empty request body", "user_id", user_id, "project_id", project_id)
		utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		h.logger.Error("[HandleUpdate] Unable to parse request", "error", pkgerr.WithStack(err), "user_id", user_id, "project_id", project_id)
		utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
		return
	}

	if _, err := body.Validate(); err != nil {
		h.logger.Error("[HandleUpdate] Unable to validate request", "error", pkgerr.WithStack(err), "user_id", user_id, "project_id", project_id)
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	new_project, err := h.project_service.UpdateOne(user_id, project_id, body.Name)

	if err != nil {
		h.logger.Error("[HandleUpdate] Unable to update project", "error", pkgerr.WithStack(err), "user_id", user_id, "project_id", project_id, "project_name", body.Name)
		switch err.Error() {
		case "invalid_role":
			utils.ResponseWithError(w, http.StatusForbidden, nil, "Insufficient permission to update project")
		default:
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
		}
		return
	}

	utils.ResponseWithSuccess(
		w,
		http.StatusOK,
		new_project,
		"Project updated successfuly",
	)
}

func (h *project_handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	project_id := r.PathValue("project_id")
	if project_id == "" {
		utils.ResponseWithSuccess[any](
			w,
			http.StatusNotFound,
			nil,
			"Project id not specified",
		)
		return
	}

	err := h.project_service.DeleteOne(user_id, project_id)

	if err != nil {
		h.logger.Error("[HandleDelete] Unable to delete project", "error", pkgerr.WithStack(err), "user_id", user_id, "project_id", project_id)
		switch err.Error() {
		case "invalid_role":
			utils.ResponseWithError(w, http.StatusForbidden, nil, "Insufficient permission to delete project")
		default:
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
		}
		return
	}

	utils.ResponseWithSuccess[any](
		w,
		http.StatusNoContent,
		nil,
		"Project updated successfuly",
	)
}