package application

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

func TestApplicationServiceListByProject(t *testing.T) {
	ctx := context.Background()
	application_repository := &StubApplicationRepository{}
	project_service := StubProjectService{}

	application_service := NewService(
		ctx,
		slog.Default(),
		application_repository,
		&project_service,
	)

	mock_user_id := "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"
	mock_project_id := "64c5e7da-3e02-4db8-aa2a-aa5161c085f7"

	t.Run("it should pass the user and project UUIDs to the repository", func (t *testing.T) {
		defer application_repository.Clear()

		application_repository.list_by_project_return = []database.ListApplicationsByProjectRow{}

		_, err := application_service.ListByProject(mock_user_id, mock_project_id)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}

		got_n_calls := application_repository.list_by_project_n_calls
		want_n_calls := 1
		if got_n_calls != want_n_calls {
			t.Fatalf("got list by project called %d times, want %d", got_n_calls, want_n_calls)
		}

		got_args := application_repository.list_by_project_call_args[0]

		got_user_id := got_args.UserID.String()
		want_user_id := mock_user_id
		if got_user_id != want_user_id {
			t.Errorf("got user_id %s, want %s", got_user_id, want_user_id)
		}

		got_project_id := got_args.ProjectID.String()
		want_project_id := mock_project_id
		if got_project_id != want_project_id {
			t.Errorf("got project_id %s, want %s", got_project_id, want_project_id)
		}
	})

	t.Run("it should return an empty list when there are no rows", func (t *testing.T) {
		defer application_repository.Clear()

		application_repository.list_by_project_error = pgx.ErrNoRows

		got, err := application_service.ListByProject(mock_user_id, mock_project_id)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
		if got == nil {
			t.Fatalf("got nil, want an empty list")
		}

		got_len := len(got)
		want_len := 0
		if got_len != want_len {
			t.Errorf("got %d entries, want %d", got_len, want_len)
		}
	})

	t.Run("it should return the generic error when the repository fails", func (t *testing.T) {
		defer application_repository.Clear()

		application_repository.list_by_project_error = errors.New("connection refused")

		got, err := application_service.ListByProject(mock_user_id, mock_project_id)
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}

		want_error := "unable to list applications, db query failed"
		if err == nil || err.Error() != want_error {
			t.Errorf("got error %v, want %s", err, want_error)
		}
	})

	t.Run("it should return one entry per application with its id, name, type and timestamps", func (t *testing.T) {
		defer application_repository.Clear()

		mock_app_id := "7aaa1bf8-437f-4f3c-8691-8316fc6fbe50"
		mock_created_at := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
		mock_updated_at := time.Date(2026, 9, 2, 11, 30, 0, 0, time.UTC)

		row := database.ListApplicationsByProjectRow{
			Name: "storefront",
			Type: "container_nodejs",
			CreatedAt: pgtype.Timestamp{Time: mock_created_at, Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: mock_updated_at, Valid: true},
		}
		row.AppID.Scan(mock_app_id)
		application_repository.list_by_project_return = []database.ListApplicationsByProjectRow{row}

		got, err := application_service.ListByProject(mock_user_id, mock_project_id)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d entries, want 1", len(got))
		}

		got_entry := got[0]
		if got_entry.AppID != mock_app_id {
			t.Errorf("got app_id %s, want %s", got_entry.AppID, mock_app_id)
		}
		if got_entry.Name != "storefront" {
			t.Errorf("got name %s, want storefront", got_entry.Name)
		}
		if got_entry.Type != "container_nodejs" {
			t.Errorf("got type %s, want container_nodejs", got_entry.Type)
		}
		if !got_entry.CreatedAt.Equal(mock_created_at) {
			t.Errorf("got created_at %v, want %v", got_entry.CreatedAt, mock_created_at)
		}
		if !got_entry.UpdatedAt.Equal(mock_updated_at) {
			t.Errorf("got updated_at %v, want %v", got_entry.UpdatedAt, mock_updated_at)
		}
	})

	t.Run("it should derive status from the current deployment instance", func (t *testing.T) {
		tests := []struct{
			name string
			instance_status pgtype.Int4
			want_status string
		}{
			{"no instance", pgtype.Int4{}, "not_deployed"},
			{"running instance", pgtype.Int4{Int32: 1, Valid: true}, "running"},
			{"stopped instance", pgtype.Int4{Int32: 0, Valid: true}, "stopped"},
			{"unknown instance value", pgtype.Int4{Int32: 7, Valid: true}, "stopped"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func (t *testing.T) {
				defer application_repository.Clear()

				application_repository.list_by_project_return = []database.ListApplicationsByProjectRow{
					{Name: "storefront", InstanceStatus: tc.instance_status},
				}

				got, err := application_service.ListByProject(mock_user_id, mock_project_id)
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				if len(got) != 1 {
					t.Fatalf("got %d entries, want 1", len(got))
				}

				got_status := got[0].Status
				if got_status != tc.want_status {
					t.Errorf("got status %s, want %s", got_status, tc.want_status)
				}
			})
		}
	})

	t.Run("it should return a null latest_deployment when the application has no deployment", func (t *testing.T) {
		defer application_repository.Clear()

		application_repository.list_by_project_return = []database.ListApplicationsByProjectRow{
			{Name: "storefront"},
		}

		got, err := application_service.ListByProject(mock_user_id, mock_project_id)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d entries, want 1", len(got))
		}

		got_latest := got[0].LatestDeployment
		if got_latest != nil {
			t.Errorf("got latest_deployment %+v, want nil", got_latest)
		}
	})

	t.Run("it should return the latest deployment summary with its derived outcome", func (t *testing.T) {
		mock_app_dp_id := "c0a1f1c6-2f7e-4c4b-9a39-0d3a1b2c3d4e"
		mock_created_at := time.Date(2026, 9, 3, 8, 0, 0, 0, time.UTC)
		mock_updated_at := time.Date(2026, 9, 3, 8, 5, 0, 0, time.UTC)

		tests := []struct{
			name string
			dp_status int32
			want_outcome string
		}{
			{"failed step", -1, "failed"},
			{"initiated", 1, "building"},
			{"build extracted", 2, "building"},
			{"image built", 3, "building"},
			{"image pushed", 4, "building"},
			{"instance started", 5, "succeeded"},
			{"unknown step above range", 6, "failed"},
			{"unknown step zero", 0, "failed"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func (t *testing.T) {
				defer application_repository.Clear()

				row := database.ListApplicationsByProjectRow{
					Name: "storefront",
					LatestDpVersionNumber: pgtype.Int4{Int32: 3, Valid: true},
					LatestDpStatus: pgtype.Int4{Int32: tc.dp_status, Valid: true},
					LatestDpCreatedAt: pgtype.Timestamp{Time: mock_created_at, Valid: true},
					LatestDpUpdatedAt: pgtype.Timestamp{Time: mock_updated_at, Valid: true},
				}
				row.LatestDpID.Scan(mock_app_dp_id)
				application_repository.list_by_project_return = []database.ListApplicationsByProjectRow{row}

				got, err := application_service.ListByProject(mock_user_id, mock_project_id)
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				if len(got) != 1 {
					t.Fatalf("got %d entries, want 1", len(got))
				}

				got_latest := got[0].LatestDeployment
				if got_latest == nil {
					t.Fatalf("got nil latest_deployment, want a summary")
				}
				if got_latest.Outcome != tc.want_outcome {
					t.Errorf("got outcome %s, want %s", got_latest.Outcome, tc.want_outcome)
				}
				if got_latest.Status != int(tc.dp_status) {
					t.Errorf("got status %d, want %d", got_latest.Status, tc.dp_status)
				}
				if got_latest.AppDpID != mock_app_dp_id {
					t.Errorf("got app_dp_id %s, want %s", got_latest.AppDpID, mock_app_dp_id)
				}
				if got_latest.VersionNumber != 3 {
					t.Errorf("got version_number %d, want 3", got_latest.VersionNumber)
				}
				if !got_latest.CreatedAt.Equal(mock_created_at) {
					t.Errorf("got created_at %v, want %v", got_latest.CreatedAt, mock_created_at)
				}
				if !got_latest.UpdatedAt.Equal(mock_updated_at) {
					t.Errorf("got updated_at %v, want %v", got_latest.UpdatedAt, mock_updated_at)
				}
			})
		}
	})

	t.Run("it should keep showing running while a new deployment rolls out", func (t *testing.T) {
		tests := []struct{
			name string
			dp_status int32
			want_outcome string
		}{
			{"latest deployment building", 2, "building"},
			{"latest deployment failed", -1, "failed"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func (t *testing.T) {
				defer application_repository.Clear()

				row := database.ListApplicationsByProjectRow{
					Name: "storefront",
					InstanceStatus: pgtype.Int4{Int32: 1, Valid: true},
					LatestDpVersionNumber: pgtype.Int4{Int32: 4, Valid: true},
					LatestDpStatus: pgtype.Int4{Int32: tc.dp_status, Valid: true},
				}
				row.LatestDpID.Scan("c0a1f1c6-2f7e-4c4b-9a39-0d3a1b2c3d4e")
				application_repository.list_by_project_return = []database.ListApplicationsByProjectRow{row}

				got, err := application_service.ListByProject(mock_user_id, mock_project_id)
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				if len(got) != 1 {
					t.Fatalf("got %d entries, want 1", len(got))
				}

				got_status := got[0].Status
				want_status := "running"
				if got_status != want_status {
					t.Errorf("got status %s, want %s", got_status, want_status)
				}

				got_latest := got[0].LatestDeployment
				if got_latest == nil {
					t.Fatalf("got nil latest_deployment, want a summary")
				}
				if got_latest.Outcome != tc.want_outcome {
					t.Errorf("got outcome %s, want %s", got_latest.Outcome, tc.want_outcome)
				}
			})
		}
	})
}
