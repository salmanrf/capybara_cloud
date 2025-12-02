package dto

import (
	"errors"

	"github.com/salmanrf/capybara-cloud/internal/database"
)

type CreateProjectDto struct {
	OrgId string `json:"org_id"`
	Name string `json:"name"`
}

type UpdateProjectDto struct {
	Name string `json:"name"`
}

type ListMyProjectEntryProject struct {
	OrgID string `json:"org_id"`
	ProjectID string `json:"project_id"`
	Name string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListMyProjectEntry struct {
	Role string `json:"role"`
	Project *ListMyProjectEntryProject `json:"project"`
}

type ListMyProjectResponse = []ListMyProjectEntry

func (dto *CreateProjectDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil
	
	if len(dto.OrgId) != 36 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("invalid org_id format, must be a valid uuid string"))
	}

	if len(dto.Name) < 5 || len(dto.Name) > 100 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("project name must have 5 to 100 characters"))
	}

	return valid, validation_errors
}

func (dto *UpdateProjectDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil
	
	if len(dto.Name) < 5 || len(dto.Name) > 100 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("project name must have 5 to 100 characters"))
	}

	return valid, validation_errors
} 

func NewGetOneProjectResponse(row database.FindOneProjectByIdRow) *ListMyProjectEntryProject {
	return &ListMyProjectEntryProject{
		OrgID: row.OrgID.String(),
		ProjectID: row.ProjectID.String(),
		Name: row.Name.String,
		CreatedAt: row.CreatedAt.Time.String(),
		UpdatedAt: row.UpdatedAt.Time.String(),
	}
}

func NewListMyProjectResponse(dbrows []database.FindProjectsForUserRow) []ListMyProjectEntry {
	size := len(dbrows)
	formatted := make([]ListMyProjectEntry, size)

	for i, en := range dbrows {
		entry := ListMyProjectEntry{
			Role: en.Role,
			Project: &ListMyProjectEntryProject{
				OrgID: en.OrgID.String(),
				ProjectID: en.ProjectID.String(),
				Name: en.Name.String,
				CreatedAt: en.ProjectCreatedAt.Time.String(),
				UpdatedAt: en.ProjectUpdatedAt.Time.String(),
			},
		}

		formatted[i] = entry
	}

	return formatted
}


