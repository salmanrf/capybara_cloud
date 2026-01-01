package dto

import (
	"bytes"
	"encoding/json"
	"errors"
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
	VariablesJSON string `json:"variables_json"`
	VariablesMap *map[string]any
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
	var variables_map map[string]any
	
	buffer := bytes.NewBuffer([]byte(dto.VariablesJSON))
	decoder := json.NewDecoder(buffer)
	if err := decoder.Decode(&variables_map); err != nil {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("invalid variables json"), err)
	}

	dto.VariablesMap = &variables_map

	return valid, validation_errors
} 

