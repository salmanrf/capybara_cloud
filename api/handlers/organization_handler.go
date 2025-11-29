package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func CreateOrgHandler(os organization.Service) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var body dto.CreateOrgDto

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

		org, err := os.Create(user_id, body.Name)
		if err != nil {
			errmsg := err.Error()
			fmt.Println("CreateOrg failed", errmsg)
			if strings.Contains(errmsg, "duplicate key") {
				utils.ResponseWithError(w, http.StatusBadRequest, nil, "Organization with this name already exists")	
			} else {
				utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
			}
			return
		}
		
		utils.ResponseWithSuccess(
			w,
			http.StatusCreated,
			org,
			"Organization created successfully",
		)
	}
}

func GetOneOrgHandler(os organization.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		org_id := strings.TrimPrefix(r.URL.Path, "/api/organizations/")
		if org_id == "" {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Organization id not specified",
			)
			return
		}

		org, err := os.FindById(user_id, org_id)

		if err != nil {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusOK,
				nil,
				err.Error(),
			)		
			return
		}

		if org == nil {
			utils.ResponseWithError(
				w,
				http.StatusNotFound,
				nil,
				"Organization not found",
			)
			return
		}

		utils.ResponseWithSuccess(
			w,
			http.StatusOK,
			org,
			"Organization retrieved successfully",
		)
	}
}

func ListMyOrgHandler(os organization.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		orguses, err := os.ListMyOrgs(user_id)
		if err != nil {
			utils.ResponseWithError(
				w,
				http.StatusInternalServerError,
				nil,
				"Internal server error",
			)
			return
		}

		formatted := dto.NewListMyOrgResponse(orguses) 

		utils.ResponseWithSuccess(
			w,
			http.StatusOK,
			&formatted,
			"Organization users retrieved successfuly",
		)
	}
}

func UpdateOneOrgHandler(os organization.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		org_id := strings.TrimPrefix(r.URL.Path, "/api/organizations/")
		if org_id == "" {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Organization id not specified",
			)
			return
		}

		var body dto.CreateOrgDto

		if r.Body == nil {
			fmt.Println("Update one organization failed, empty body")
			utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
			return
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&body); err != nil {
			fmt.Println("Update one organization failed, ", err.Error())
			utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
			return
		}

		if _, err := body.Validate(); err != nil {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
		}

		org, err := os.FindByIdAndRole(
			user_id,
			org_id,
			[]string{"owner"},
		)

		if err != nil {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Organization not found",
			)
			return
		}

		if org == nil {
			utils.ResponseWithError(
				w,
				http.StatusNotFound,
				nil,
				"Organization not found",
			)
			return
		}

		if org.Role != "owner" { 
			utils.ResponseWithError(
				w,
				http.StatusForbidden,
				nil,
				"Insufficient permission to update organization",
			)
			return
		}

		org.Name = pgtype.Text{String: body.Name, Valid: true}

		new_org, err := os.UpdateOne(org)
		
		if err != nil {
			fmt.Println("Update one org failed, ", err)
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
			new_org,
			"Organization updated successfuly",
		)
	}
}

func DeleteOneOrgHandler(os organization.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()
		user_id := rctx.Value("user_id").(string)

		org_id := strings.TrimPrefix(r.URL.Path, "/api/organizations/")
		if org_id == "" {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Organization id not specified",
			)
			return
		}

		org, err := os.FindByIdAndRole(
			user_id,
			org_id,
			[]string{"owner"},
		)

		if err != nil {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNotFound,
				nil,
				"Organization not found",
			)
			return
		}

		if org == nil {
			utils.ResponseWithSuccess[any](
				w,
				http.StatusNoContent,
				nil,
				"Organization updated successfuly",
			)
			return
		}

		if org.Role != "owner" { 
			utils.ResponseWithError(
				w,
				http.StatusForbidden,
				nil,
				"Insufficient permission to update organization",
			)
			return
		}

		err = os.DeleteOne(org.OrgID.String())

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
			"Organization updated successfuly",
		)
	}
}