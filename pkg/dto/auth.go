package dto

import "github.com/salmanrf/capybara-cloud/internal/database"

type AuthMeResponse struct {
	UserId string `json:"user_id"`
	Email string `json:"email"`
	FullName string `json:"full_name"`
}

func NewAuthMeResponse(user *database.User) *AuthMeResponse {
	return &AuthMeResponse{
		UserId: user.UserID.String(),
		Email: user.Email,
		FullName: user.FullName,
	}
}

