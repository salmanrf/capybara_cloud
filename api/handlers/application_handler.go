package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

func CreateApplicationHandler(app_service application.Service) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
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
		
		new_application, err := app_service.Create(
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
}
