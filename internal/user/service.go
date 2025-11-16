package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/auth"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
)

type Service interface {
	FindById(identifier string, is_email bool) (*database.User, error)
	Create(create_params dto.SignupDto) (*database.User, error) 
}

type service struct {
	ctx context.Context
	queries *database.Queries
}

func NewService(ctx context.Context, q *database.Queries) Service {
	return &service{
		ctx: ctx,
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
		err_msg := err.Error()
			if strings.Contains(err_msg, "no rows") {
				return nil, nil
			} else {
				fmt.Println("Error finding user", err_msg)
				return nil, errors.New("unable to find user")
			}
	}

	return &user, nil
}

func (s *service) Create(dto dto.SignupDto) (*database.User, error) {
	hashed_password, err := auth.Hash(dto.Password)

	if err != nil {
		fmt.Println("Error at user_service.Create - hashing user password: ", err.Error())
		return nil, err
	}

	user, err := s.queries.CreateOneUser(s.ctx, database.CreateOneUserParams{
		Email: dto.Email,
		Username: dto.Username,
		FullName: dto.FullName,
		HashedPassword: hashed_password,
	})

	if err != nil {
		errr_msg := err.Error()
		fmt.Println("Error at user_service.Create - inserting to db: ", errr_msg)
		return nil, err
	}

	return &user, nil
}