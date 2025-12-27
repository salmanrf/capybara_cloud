package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

type project_handler struct {
	project_service project.Service
}

type ProjectHandlers interface {
	HandleCreate(w http.ResponseWriter, r *http.Request)
	HandleFindOne(w http.ResponseWriter, r *http.Request)
	HandleListMyProjects(w http.ResponseWriter, r *http.Request)
	HandleUpdate(w http.ResponseWriter, r *http.Request)
	HandleDelete(w http.ResponseWriter, r *http.Request)
}

func NewProjectHandlers(project_service project.Service) ProjectHandlers {
	return &project_handler{
		project_service,
	}
}

func (h *project_handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var body dto.CreateProjectDto

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
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
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	project, err := h.project_service.Create(user_id, body.OrgId, body.Name)
	if err != nil {
		errmsg := err.Error()
		fmt.Println("CreateProject failed", errmsg)
		if strings.Contains(errmsg, "duplicate key") {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Project with this name already exists")
		} else {
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

	projectuses, err := h.project_service.ListMyProjects(user_id)
	if err != nil {
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
		fmt.Println("Update one project failed, empty body")
		utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		fmt.Println("Update one project failed, ", err.Error())
		utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
		return
	}

	if _, err := body.Validate(); err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
	}

	project, err := h.project_service.FindByIdAndRole(
		user_id,
		project_id,
		[]string{"owner"},
	)

	if err != nil {
		utils.ResponseWithSuccess[any](
			w,
			http.StatusNotFound,
			nil,
			"Project not found",
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

	if project.Role != "owner" {
		utils.ResponseWithError(
			w,
			http.StatusForbidden,
			nil,
			"Insufficient permission to update projectanization",
		)
		return
	}

	project.Name = pgtype.Text{String: body.Name, Valid: true}

	new_project, err := h.project_service.UpdateOne(project)

	if err != nil {
		fmt.Println("Update one project failed, ", err)
		utils.ResponseWithSuccess[any](
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

	project, err := h.project_service.FindByIdAndRole(
		user_id,
		project_id,
		[]string{"owner"},
	)

	if err != nil {
		utils.ResponseWithSuccess[any](
			w,
			http.StatusNotFound,
			nil,
			"Project not found",
		)
		return
	}

	if project == nil {
		utils.ResponseWithSuccess[any](
			w,
			http.StatusNoContent,
			nil,
			"Project updated successfuly",
		)
		return
	}

	if project.Role != "owner" {
		utils.ResponseWithError(
			w,
			http.StatusForbidden,
			nil,
			"Insufficient permission to update projectanization",
		)
		return
	}

	err = h.project_service.DeleteOne(project.ProjectID.String())

	if err != nil {
		utils.ResponseWithError(
			w,
			http.StatusInternalServerError,
			nil,
			"Internal server error",
		)
		return
	}

	utils.ResponseWithSuccess[any](
		w,
		http.StatusNoContent,
		nil,
		"Project updated successfuly",
	)
}