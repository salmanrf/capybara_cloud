package organization

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/user"
)

type Service interface {
	Create(user_id string, org_name string) (*database.Organization, error)
	UpdateOne(dto *database.FindOneOrganizationByIdAndRoleRow) (*database.Organization, error)
	DeleteOne(org_id string) error
	FindById(user_id string, org_id string) (*database.FindOneOrganizationByIdRow, error)
	FindByIdAndRole(user_id string, org_id string, roles []string) (*database.FindOneOrganizationByIdAndRoleRow, error)
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

func (s *service) UpdateOne(dto *database.FindOneOrganizationByIdAndRoleRow) (*database.Organization, error) {
	updated_at := pgtype.Timestamp{
		Time: time.Now(),
		Valid: true,
	}
	
	organization, err := s.queries.UpdateOneOrganization(s.ctx, database.UpdateOneOrganizationParams{
		OrgID: dto.OrgID,
		Name: dto.Name.String,
		UpdatedAt: updated_at,
	})

	if err != nil {
		fmt.Println("Error updating organization", organization.OrgID.String(), organization.Name, err)
		return nil, err
	}

	return &organization, nil
}

func (s *service) DeleteOne(org_id string) error {
	org_uuid := pgtype.UUID{}
	org_uuid.Scan(org_id)

	err := s.queries.DeleteOneOrganization(s.ctx, org_uuid)

	return err
}

func (s *service) FindById(user_id string, org_id string) (*database.FindOneOrganizationByIdRow, error) {
	org_uuid := pgtype.UUID{}
	org_uuid.Scan(org_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	org_res, err := s.queries.FindOneOrganizationById(s.ctx, database.FindOneOrganizationByIdParams{
		OrgID: org_uuid,
		UserID: user_uuid,
	})

	if err != nil {
		err_msg := err.Error()
			if strings.Contains(err_msg, "no rows") {
				return nil, nil
			} else {
				fmt.Println("Error finding organization", err_msg)
				return nil, errors.New("unable to find organization")
			}
	}

	return &org_res, nil
}

func (s *service) FindByIdAndRole(user_id string, org_id string, roles []string) (*database.FindOneOrganizationByIdAndRoleRow, error) {
	org_uuid := pgtype.UUID{}
	org_uuid.Scan(org_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	org_res, err := s.queries.FindOneOrganizationByIdAndRole(s.ctx, database.FindOneOrganizationByIdAndRoleParams{
		OrgID: org_uuid,
		UserID: user_uuid,
	})

	if err != nil {
		err_msg := err.Error()
			if strings.Contains(err_msg, "no rows") {
				return nil, nil
			} else {
				fmt.Println("Error finding user", err_msg)
				return nil, errors.New("unable to find user")
			}
	}

	return &org_res, nil
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