package organization

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/user"
)

type Service interface {
	Create(user_id string, org_name string) (*database.Organization, error)
	ListMyOrgs(user_id string) ([]database.FindOrganizationsForUserRow, error)
}

type service struct {
	ctx context.Context
	queries *database.Queries
	user_service user.Service
}

func NewService(ctx context.Context, queries *database.Queries, user_service user.Service) Service {
	return &service{
		ctx,
		queries,
		user_service,
	}
}

func (s *service) Create(user_id string, org_name  string) (*database.Organization, error) {
	user, err := s.user_service.FindById(user_id, false)

	if err != nil {
		return nil, err
	}
	
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

func (s *service) ListMyOrgs(user_id string) ([]database.FindOrganizationsForUserRow, error) {
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	orgus, err := s.queries.FindOrganizationsForUser(s.ctx, user_uuid)
	if err != nil {
		errmsg := err.Error()		
		fmt.Println("Error at organization_service.ListMyOrgs", errmsg)
		if strings.Contains(errmsg, "no rows") {
			return []database.FindOrganizationsForUserRow{}, nil
		}
		return nil, errors.New("unable to find organization users, db query failed")
	}

	return orgus, nil
}