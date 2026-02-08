package application

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/tests"
)

func TestApplicationService(t *testing.T) {
	ctx := context.Background()
	pgxpool := &pgxpool.Pool{}
	application_repository := &StubApplicationRepository{}
	project_service := tests.StubProjectService{}

	application_service := NewService(
		ctx,
		pgxpool,
		application_repository,
		&project_service,
	)

	t.Run("should return error not_found when app_with_pm returns nil", func (t *testing.T) {
		defer application_repository.Clear()
		
		app_id := "a7e4e583-471c-4b51-bcdd-7fb57291c5cb"
		user_id := "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"

		application_repository.find_one_with_project_member_return = nil

		_, err := application_service.CreateConfig(
			app_id, 
			user_id,
			dto.CreateApplicationConfigDto{},
		)		

		got_error := err
		want_error := errors.New("not_found")

		if err == nil || got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})

	t.Run("should return error permission_denied when there's no matching project member", func (t *testing.T) {
		defer application_repository.Clear()

		app_id := "a7e4e583-471c-4b51-bcdd-7fb57291c5cb"
		user_id := "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"

		mock_app_with_pm := &database.FindOneApplicationWithProjectMemberRow{}
		mock_app_with_pm.AppID.Valid = true
		mock_app_with_pm.PmProjectID.Valid = false
		application_repository.find_one_with_project_member_return = mock_app_with_pm
		
		_, err := application_service.CreateConfig(
			app_id, 
			user_id,
			dto.CreateApplicationConfigDto{},
		)		

		got_error := err
		want_error := errors.New("permission_denied")

		if err == nil || got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})
}