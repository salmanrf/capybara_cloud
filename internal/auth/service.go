package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/user"
)

type Service interface {
	GetMe(user_id string) (*database.User, error)
}

type service struct {
	ctx context.Context
	user_service user.Service
}

func NewService(ctx context.Context, user_service user.Service) Service {
	return &service{
		ctx,
		user_service,
	}
}

func (s *service) GetMe(user_id string) (*database.User, error) {
	user, err := s.user_service.FindById(user_id, false)

	if err != nil {
		fmt.Println("Error at auth_service.GetMe", err)
		return nil, errors.New("unable to find user")
	}

	return user, nil
}