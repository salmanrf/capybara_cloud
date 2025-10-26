package handlers

import (
	"net/http"

	"github.com/salmanrf/capybara-cloud/internal/organization"
)

func CreateOrgHandler(s organization.Service) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {

	}
}