package dto

import (
	"errors"

	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

type SigninDto struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func (dto *SigninDto) Validate() (bool, error) {
	valid := true
	var validation_errors error = nil
	
	if valid := utils.ValidateEmail(dto.Email, 100); !valid {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("invalid email address"))
	}

	if valid := utils.ValidatePassword(dto.Password); !valid {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("passwords must have at least 8 characters, one uppercase letter, one symbol, and one number"))
	}

	return valid, validation_errors
}