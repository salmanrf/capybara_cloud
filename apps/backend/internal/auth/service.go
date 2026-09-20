package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/salmanrf/capybara-cloud/apps/backend/internal/user"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type Service interface {
	GetMe(user_id string) (*database.User, error)
}

type service struct {
	ctx context.Context
	logger *slog.Logger
	user_service user.Service
}

func NewService(ctx context.Context, logger *slog.Logger, user_service user.Service) Service {
	return &service{
		ctx,
		logger,
		user_service,
	}
}

func (s *service) GetMe(user_id string) (*database.User, error) {
	user, err := s.user_service.FindById(user_id, false)

	if err != nil {
		return nil, errors.New("unable to find user")
	}

	return user, nil
}