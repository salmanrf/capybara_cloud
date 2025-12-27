package dto

import (
	"errors"
	"slices"
)

func GetSupportedAppTypes() []string {
	return []string{
		"web_app_container",
	}
}

type CreateApplicationDto struct {
	ProjectID string `json:"project_id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type UpdateApplicationDto struct {
	Name string `json:"name"`
}

type ListMyApplicationEntryApplication struct {
	AppID string `json:"app_id"`
	ProjectID string `json:"project_id"`
	Type string `json:"type"`
	Name string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListMyApplicationResponse = []ListMyApplicationEntryApplication

func (dto *CreateApplicationDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil
	
	if len(dto.ProjectID) != 36 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("invalid project_id format, must be a valid uuid string"))
	}

	if len(dto.Name) < 5 || len(dto.Name) > 100 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("app name must have 5 to 100 characters"))
	}

	if !slices.Contains(GetSupportedAppTypes(), dto.Type) {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("application type not supported"))
	}

	return valid, validation_errors
}

func (dto *UpdateApplicationDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil
	
	if len(dto.Name) < 5 || len(dto.Name) > 100 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("project name must have 5 to 100 characters"))
	}

	return valid, validation_errors
} 


