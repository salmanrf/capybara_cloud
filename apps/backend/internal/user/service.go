package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/auth"
	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

type Service interface {
	FindById(identifier string, is_email bool) (*database.User, error)
	Create(create_params dto.SignupDto) (*database.User, error) 
}

type service struct {
	ctx context.Context
	logger *slog.Logger
	queries *database.Queries
}

func NewService(ctx context.Context, logger *slog.Logger, q *database.Queries) Service {
	return &service{
		ctx: ctx,
		logger: logger,
		queries: q,
	}
}

func (s *service) FindById(identifier string, is_email bool) (*database.User, error) {
	var user database.User
	var err error
	
	if is_email {
		user, err = s.queries.FindOneUserByEmail(s.ctx, identifier)
	} else {
		user_uuid := pgtype.UUID{}
		user_uuid.Scan(identifier)

		user, err = s.queries.FindOneUserById(s.ctx, user_uuid)
	}

	if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, nil
			} else {
				return nil, errors.New("unable to find user")
			}
	}

	return &user, nil
}

func (s *service) Create(dto dto.SignupDto) (*database.User, error) {
	hashed_password, err := auth.Hash(dto.Password)

	if err != nil {
		return nil, err
	}

	user, err := s.queries.CreateOneUser(s.ctx, database.CreateOneUserParams{
		Email: dto.Email,
		Username: dto.Username,
		FullName: dto.FullName,
		HashedPassword: hashed_password,
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}