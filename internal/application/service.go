package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/project"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
)

type Service interface {
	Create(user_id string, dto dto.CreateApplicationDto) (*database.Application, error)
	Update(app_id string, user_id string, dto dto.UpdateApplicationDto) (*database.Application, error)
	FindOne(app_id string, user_id string) (*database.FindOneApplicationWithProjectMemberRow, error)
	CreateConfig(app_id string, user_id string, dto dto.CreateApplicationConfigDto) (*database.ApplicationConfig, error)
}

type service struct {
	ctx context.Context
	conn *pgxpool.Pool
	repository ApplicationRepository
	project_service project.Service
}

func NewService(
	ctx context.Context, 
	conn *pgxpool.Pool, 
	repository ApplicationRepository,
	project_service project.Service,
) Service {
	return &service{
		ctx,
		conn,
		repository,
		project_service,
	}
}

func (s *service) Create(user_id string, dto dto.CreateApplicationDto) (*database.Application, error) {
	proj_user, err :=  s.project_service.FindByIdAndRole(user_id, dto.ProjectID, []string{"member"})

	if err != nil {
		return nil, err
	}

	if proj_user == nil {
		permission_err := errors.New("permisssion_denied")
		return nil, permission_err
	}

	project_uuid := pgtype.UUID{}
	project_uuid.Scan(dto.ProjectID)

	params := database.CreateApplicationParams{
		Type: dto.Type,
		ProjectID: project_uuid,
		Name: dto.Name,
	}

	new_application, err := s.repository.CreateApplication(params)
	if err != nil {
		return nil, err
	}

	return new_application, nil
}

func (s *service) FindOne(app_id string, user_id string) (*database.FindOneApplicationWithProjectMemberRow, error) {
	app_uuid := pgtype.UUID{}
	app_uuid.Scan(app_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	app_with_pm, err := s.repository.FindOneWithProjectMember(
		database.FindOneApplicationWithProjectMemberParams{
			AppID: app_uuid,
			UserID: user_uuid,
		},
	)

	if !app_with_pm.AppID.Valid {
		return nil, nil
	}

	if !app_with_pm.PmProjectID.Valid {
		return nil, errors.New("permission_denied")
	}
	
	return app_with_pm, err
}

func (s *service) Update(app_id string, user_id string, dto dto.UpdateApplicationDto) (*database.Application, error) {
	app_with_pm, err := s.FindOne(app_id, user_id)

	if err != nil {
		errmsg := err.Error()
		if strings.Contains(errmsg, "no rows") {
			return nil, errors.New("not_found")
		}
		return nil, err
	}

	if app_with_pm == nil {
		return nil, errors.New("not_found")
	}

	updated_app, err := s.repository.UpdateOneApplication(
		database.UpdateOneApplicationParams{
			AppID: app_with_pm.AppID,
			Name: dto.Name,
			UpdatedAt: pgtype.Timestamp{
				Time: time.Now(),
				Valid: true,
			},
		},
	)

	if err != nil {
		return nil, err
	}

	return updated_app, nil
}

func (s *service) CreateConfig(app_id string, user_id string, dto dto.CreateApplicationConfigDto) (*database.ApplicationConfig, error) {
	app_uuid := pgtype.UUID{}
	app_uuid.Scan(app_id)
	user_uuid := pgtype.UUID{}
	user_uuid.Scan(user_id)

	app_with_pm, err := s.repository.FindOneWithProjectMember(
		database.FindOneApplicationWithProjectMemberParams{
			AppID: app_uuid,
			UserID: user_uuid,
		},
	)

	if err != nil {
		return nil, err
	}

	if app_with_pm == nil || !app_with_pm.AppID.Valid {
		return nil, nil
	}
	if !app_with_pm.PmProjectID.Valid {
		return nil, errors.New("permission_denied")
	}

	variables_json := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(variables_json)
	if err := encoder.Encode(dto.Variables); err != nil {
		return nil, err
	}

	params := database.CreateApplicationConfigParams{
		AppID: app_uuid,
		VariablesJson: variables_json.Bytes(),
	} 
	app_cfg, err := s.repository.CreateConfig(params)

	return app_cfg, nil
}