package dto

import (
	"errors"

	"github.com/salmanrf/capybara-cloud/internal/database"
)

type CreateOrgDto struct {
	Name string `json:"name"`
}

type ListMyOrgEntryOrg struct {
	OrgID string `json:"org_id"`
	Name string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListMyOrgEntry struct {
	Role string `json:"role"`
	Organization *ListMyOrgEntryOrg `json:"organization"`
}

type ListMyOrgResponse = []ListMyOrgEntry

func (dto *CreateOrgDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil
	
	if len(dto.Name) < 5 || len(dto.Name) > 100 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("organization name must have 5 to 100 characters"))
	}

	return valid, validation_errors
} 

func NewListMyOrgResponse(dbrows []database.FindOrganizationsForUserRow) []ListMyOrgEntry {
	size := len(dbrows)
	formatted := make([]ListMyOrgEntry, size)

	for i, en := range dbrows {
		entry := ListMyOrgEntry{
			Role: en.Role,
			Organization: &ListMyOrgEntryOrg{
				OrgID: en.OrgID.String(),
				Name: en.Name.String,
				CreatedAt: en.OrgCreatedAt.Time.String(),
				UpdatedAt: en.OrgUpdatedAt.Time.String(),
			},
		}

		formatted[i] = entry
	}

	return formatted
}


