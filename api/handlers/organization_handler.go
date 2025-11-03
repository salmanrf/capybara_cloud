package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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