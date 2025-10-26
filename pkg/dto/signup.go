package dto

import (
	"errors"
	"regexp"

	"github.com/salmanrf/capybara-cloud/pkg/utils"
)

type SignupDto struct {
	Email string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

func (dto *SignupDto) Validate() (bool, error) {
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

	if len(dto.Username) < 4 || len(dto.Username) > 100 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("username must be minimum 4 and maximum 100 characters long"))
	}

	if dto.FullName == "" || len(dto.FullName) > 250 {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("full name can't be longer than 250 characters and can't contain symbols and numbers"))
	}

	match, err := regexp.Match(`^[a-zA-Z\s]{1,250}$`, []byte(dto.FullName))

	if err != nil {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("full name can't be longer than 250 characters and can't contain symbols and numbers"))
	}

	if !match {
		valid = false
		validation_errors = errors.Join(validation_errors, errors.New("full name can't be longer than 250 characters and can't contain symbols and numbers")) 
	}

	return valid, validation_errors
}