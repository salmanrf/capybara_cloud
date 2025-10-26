package organization

import (
	"context"
	"fmt"

	"github.com/salmanrf/capybara-cloud/internal/database"
)

type Service interface {
	Create(user *database.User, org_name string) (*database.Organization, error)
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

func (s *service) Create(user *database.User, org_name  string ) (*database.Organization, error) {
	organization, err := s.queries.CreateOrganization(s.ctx, org_name)

	if err != nil {
		fmt.Println("Error creating organization", user.UserID, org_name, err)
		return nil, err
	}

	_, err = s.queries.CreateOrganizationUser(
		s.ctx, 
		database.CreateOrganizationUserParams{
			OrgID: organization.OrgID,
			UserID: user.UserID,
			Role: "owner",
		},
	)

	if err != nil {
		fmt.Println("Error creating organization", user.UserID, org_name, err)
		return nil, err
	}
	
	return &organization, nil
}