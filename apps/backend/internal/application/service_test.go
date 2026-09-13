package application

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/dto"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

func TestApplicationService(t *testing.T) {
	ctx := context.Background()
	application_repository := &StubApplicationRepository{}
	project_service := StubProjectService{}

	application_service := NewService(
		ctx,
		application_repository,
		&project_service,
	)

	t.Run("should return error not_found when app_with_pm returns nil", func (t *testing.T) {
		defer application_repository.Clear()
		
		app_id := "a7e4e583-471c-4b51-bcdd-7fb57291c5cb"
		user_id := "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"

		application_repository.find_one_complete_return = nil

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

		mock_app_with_pm := &database.FindOneApplicationCompleteRow{}
		mock_app_with_pm.AppID.Valid = true
		mock_app_with_pm.ProjectMember.ProjectID.Valid = false
		application_repository.find_one_complete_return = mock_app_with_pm
		
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

	t.Run("FindOne should return repository error instead of permission_denied", func (t *testing.T) {
		defer application_repository.Clear()

		app_id := "a7e4e583-471c-4b51-bcdd-7fb57291c5cb"
		user_id := "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"

		partial_row := &database.FindOneApplicationCompleteRow{}
		partial_row.AppID.Valid = true
		application_repository.find_one_complete_return = partial_row
		application_repository.find_one_complete_error = errors.New("can't scan into dest[11]: cannot scan NULL into *int32")

		_, err := application_service.FindOneComplete(app_id, user_id)

		if err == nil || err.Error() == "permission_denied" {
			t.Errorf("got error %v, want scan error", err)
		}
	})

	t.Run("FindOne should return nil, nil on pgx.ErrNoRows", func (t *testing.T) {
		defer application_repository.Clear()

		application_repository.find_one_complete_error = pgx.ErrNoRows

		got, err := application_service.FindOneComplete("a7e4e583-471c-4b51-bcdd-7fb57291c5cb", "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4")

		if got != nil || err != nil {
			t.Errorf("got %v, %v; want nil, nil", got, err)
		}
	})
}
