package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/salmanrf/capybara-cloud/apps/backend/internal/deployment"
	config "github.com/salmanrf/capybara-cloud/apps/backend/pkg/utils"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/utils"
)

type deployment_handler struct {
	deployment_service deployment.Service
}

type AppDeploymentHandlers interface {
	HandleCreateOneDeployment(w http.ResponseWriter, r *http.Request) 
}

func NewAppDeploymentHandlers(deployment_service deployment.Service) AppDeploymentHandlers {
	return &deployment_handler{
		deployment_service,
	}
}

func (h *deployment_handler) HandleCreateOneDeployment(w http.ResponseWriter, r *http.Request) {
	content_type := r.Header.Get("Content-Type");
	if !strings.HasPrefix(content_type, "multipart/form-data") {
		utils.ResponseWithError(
			w,
			http.StatusBadRequest,
			nil,
			fmt.Sprintf("content-type must be %s", "multipart/form-data"),
		)
		return
	}

	content_length, err := strconv.Atoi(r.Header.Get("Content-Length"));
	if err != nil || content_length == 0 {
		utils.ResponseWithError(
			w,
			http.StatusBadRequest,
			nil,
			"form-data mustn't be empty",
		)
		return
	}
	cfg := config.GetConfig()
	if content_length > cfg.MAX_DEPLOY_FORM_SIZE {
		utils.ResponseWithError(
			w,
			http.StatusRequestEntityTooLarge,
			nil,
			fmt.Sprintf("form-data payload exceeds %d in size", cfg.MAX_DEPLOY_FORM_SIZE),
		)
		return
	}

	err = r.ParseMultipartForm(int64(cfg.MAX_DEPLOY_FORM_SIZE))
	if err != nil {
		errmsg := err.Error()
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			errmsg,
		)
		return
	}

	bundle_file, file_headers, err := r.FormFile("bundle")
	if err != nil {
		errmsg := err.Error()
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			errmsg,
		)
		return
	}
	if bundle_file == nil || file_headers == nil {
		utils.ResponseWithError(
			w,
			http.StatusUnprocessableEntity,
			nil,
			"deployment 'bundle' file must be provided as .tar.gz (application/gzip) file",
		)
		return
	}
	if file_headers.Size > int64(cfg.MAX_DEPLOY_BUNDLE_SIZE) {
		utils.ResponseWithError(
			w,
			http.StatusRequestEntityTooLarge,
			nil,
			fmt.Sprintf("form-data bundle file exceeds %d in size", cfg.MAX_DEPLOY_BUNDLE_SIZE),
		)
		return
	}

	app_id := r.PathValue("app_id")
	user_id, _ := r.Context().Value("user_id").(string)

	deployment, err := h.deployment_service.Deploy(user_id, app_id, file_headers, bundle_file)
	if err != nil {
		fmt.Println("Error: ", err)
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
		http.StatusCreated,
		deployment,
		"Deployment created successfully",
	)
} 