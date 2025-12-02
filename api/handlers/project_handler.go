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

func CreateProjectHandler(ps project.Service) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
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

		project, err := ps.Create(user_id, body.OrgId, body.Name)
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
}

func GetOneProjectHandler(ps project.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		project_id := strings.TrimPrefix(r.URL.Path, "/api/projects/")
		if project_id == "" {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Project id not specified",
			)
			return
		}

		project, err := ps.FindById(user_id, project_id)

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
}

func ListMyProjectHandler(ps project.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		projectuses, err := ps.ListMyProjects(user_id)
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
}

func UpdateOneProjectHandler(ps project.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		project_id := strings.TrimPrefix(r.URL.Path, "/api/projects/")
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

		project, err := ps.FindByIdAndRole(
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

		new_project, err := ps.UpdateOne(project)
		
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
}

func DeleteOneProjectHandler(ps project.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		project_id := strings.TrimPrefix(r.URL.Path, "/api/projects/")
		if project_id == "" {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Project id not specified",
			)
			return
		}

		project, err := ps.FindByIdAndRole(
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

		err = ps.DeleteOne(project.ProjectID.String())

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
}