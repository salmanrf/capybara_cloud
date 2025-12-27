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

type org_handler struct {
	org_service organization.Service
}

type OrgHandlers interface {
	HandleFindOne(w http.ResponseWriter, r *http.Request) 
	HandleCreate(w http.ResponseWriter, r *http.Request)
	HandleUpdate(w http.ResponseWriter, r *http.Request)
	HandleDelete(w http.ResponseWriter, r *http.Request)
	HandleListMyOrganizations(w http.ResponseWriter, r *http.Request)
}

func NewOrgHandlers(org_service organization.Service) OrgHandlers {
	return &org_handler{
		org_service,
	}
}

func (h *org_handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
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

		org, err := h.org_service.Create(user_id, body.Name)
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

func (h *org_handler) HandleFindOne(w http.ResponseWriter, r *http.Request){
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	org_id := r.PathValue("org_id")

	org, err := h.org_service.FindById(user_id, org_id)

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

func (h *org_handler) HandleListMyOrganizations(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	orguses, err := h.org_service.ListMyOrgs(user_id)
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

func (h *org_handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	org_id := r.PathValue("org_id")

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

	org, err := h.org_service.FindByIdAndRole(
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

	new_org, err := h.org_service.UpdateOne(org)
	
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

func (h *org_handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	user_id := rctx.Value("user_id").(string)

	org_id := r.PathValue("org_id")

	org, err := h.org_service.FindByIdAndRole(
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

	err = h.org_service.DeleteOne(org.OrgID.String())

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