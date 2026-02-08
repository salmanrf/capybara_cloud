package dto

import (
	"errors"
	"fmt"
	"slices"
	"time"
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

type CreateApplicationConfigDto struct {
	Variables map[string]any `json:"variables"`
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

type ApplicationConfigResponse struct {
	AppCfgID string`json:"app_cfg_id"`
	AppID string `json:"app_id"`
	VariablesJson string `json:"variables_json"`
	ConfigVariables map[string]any `json:"config_variables"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

func (dto *CreateApplicationConfigDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil

	keyc := 0
	for key, val := range dto.Variables {
		switch t := val.(type) {
		case string, float32, float64, int:
			keyc += 1
		default:
			validation_errors = errors.Join(
				validation_errors,
				fmt.Errorf("%s is not a primitive data type: (%v) %v", key, t, val),
			)
			valid = false
		}
	}
	if keyc == 0 {
		validation_errors = errors.Join(
			validation_errors, 
			errors.New("config variables can't be an empty map"),
		)
		valid = false
	}
	
	return valid, validation_errors
} 

