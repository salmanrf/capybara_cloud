package project

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

const test_user_id = "3ad11d5d-5a7e-433d-ac51-fba7a645f3d4"
const test_org_id = "9b0c7f0e-6a51-4f43-9d0b-2f7b3a0f5c11"
const test_project_id = "a7e4e583-471c-4b51-bcdd-7fb57291c5cb"

func TestProjectServiceCreate(t *testing.T) {
	ctx := context.Background()
	project_repository := &StubProjectRepository{}
	user_service := &StubUserService{}
	org_service := &StubOrgService{}

	project_service := NewService(ctx, slog.Default(), project_repository, user_service, org_service)

	t.Run("should return the user lookup error without touching the repository", func(t *testing.T) {
		defer project_repository.Clear()
		defer user_service.Clear()
		defer org_service.Clear()

		user_service.find_by_id_err = errors.New("db down")

		_, err := project_service.Create(test_user_id, test_org_id, "my-project")

		got_error := err
		want_error := errors.New("db down")
		if err == nil {
			t.Fatalf("got error nil, want %v", want_error)
		}
		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}

		got_n_calls := project_repository.create_project_with_member_n_calls
		want_n_calls := 0
		if got_n_calls != want_n_calls {
			t.Errorf("got %d CreateProjectWithMember calls, want %d", got_n_calls, want_n_calls)
		}
	})

	t.Run("should return error when org_service.FindByIdAndRole returns error", func (t *testing.T) {
		defer project_repository.Clear()
		defer user_service.Clear()
		defer org_service.Clear()

		want_error := errors.New("internal error")
		org_service.find_by_id_and_role_err = want_error

		user_service.find_by_id_return = &database.User{
			UserID: test_uuid(test_user_id),
		}

		_, got_err := project_service.Create(
			test_user_id,
			test_project_id,
			"Masbro Cloud",
		)

		if got_err == nil {
			t.Fatalf("got error nil, want '%v'", want_error)
		}

		got_find_org_called := org_service.find_by_id_and_role_n_calls
		want_find_org_called := 1
		if got_find_org_called != want_find_org_called {
			t.Errorf("got org_service.FindByIdAndRole called %d time, want %d", got_find_org_called, want_find_org_called)
		}
	})

	t.Run("should return error 'invalid_role' org not found", func (t *testing.T) {
		defer project_repository.Clear()
		defer user_service.Clear()
		defer org_service.Clear()

		org_service.find_by_id_and_role_return = nil
		org_service.find_by_id_and_role_err = nil

		user_service.find_by_id_return = &database.User{
			UserID: test_uuid(test_user_id),
		}

		_, got_err := project_service.Create(
			test_user_id,
			test_project_id,
			"Masbro Cloud",
		)
		want_error := errors.New("invalid_role")

		if got_err == nil {
			t.Fatalf("got error nil, want '%v'", want_error)
		}

		got_find_org_called := org_service.find_by_id_and_role_n_calls
		want_find_org_called := 1
		if got_find_org_called != want_find_org_called {
			t.Errorf("got org_service.FindByIdAndRole called %d time, want %d", got_find_org_called, want_find_org_called)
		}
	})

	t.Run("should reject creation with 'invalid_role' when role within org is incorrect", func (t *testing.T) {
		tests := []struct{
			role string
		}{
			{"viewer"},
			{"kambing"},
			{"kucing"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("role: '%s'", tt.role), func (t *testing.T) {
				defer project_repository.Clear()
				defer user_service.Clear()
				defer org_service.Clear()

				org_service.find_by_id_and_role_return = 
					&database.FindOneOrganizationByIdAndRoleRow{
						Role: tt.role,
					}
				org_service.find_by_id_and_role_err = nil

				user_service.find_by_id_return = &database.User{
					UserID: test_uuid(test_user_id),
				}

				_, got_err := project_service.Create(
					test_user_id,
					test_project_id,
					"Masbro Cloud",
				)
				want_error := errors.New("invalid_role")

				if got_err == nil {
					t.Fatalf("got error nil, want '%v'", want_error)
				}

				got_find_org_called := org_service.find_by_id_and_role_n_calls
				want_find_org_called := 1
				if got_find_org_called != want_find_org_called {
					t.Errorf("got org_service.FindByIdAndRole called %d time, want %d", got_find_org_called, want_find_org_called)
				}
			})
		}
	})
	
	t.Run("should create the project with the user as owner", func(t *testing.T) {
		tests := []struct{
			user string
			role string
		}{
			{test_user_id, "owner"},
			{test_user_id, "editor"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("user: '%s', role: '%s'", tt.user, tt.role), func(t *testing.T) {
				defer project_repository.Clear()
				defer org_service.Clear()
				defer user_service.Clear()

				user_service.find_by_id_return = &database.User{UserID: test_uuid(test_user_id)}
				org_service.find_by_id_and_role_return = 
					&database.FindOneOrganizationByIdAndRoleRow{
						Role: tt.role,
					}
				org_service.find_by_id_and_role_err = nil
				created := &database.Project{ProjectID: test_uuid(test_project_id), Name: "my-project"}
				project_repository.create_project_with_member_return = created
		
				got_project, err := project_service.Create(test_user_id, test_org_id, "my-project")
		
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				if got_project != created {
					t.Errorf("got project %v, want %v", got_project, created)
				}
		
				got_n_calls := project_repository.create_project_with_member_n_calls
				want_n_calls := 1
				if got_n_calls != want_n_calls {
					t.Fatalf("got %d CreateProjectWithMember calls, want %d", got_n_calls, want_n_calls)
				}
		
				got_args := project_repository.create_project_with_member_call_args[0]
				want_project_params := database.CreateProjectParams{OrgID: test_uuid(test_org_id), Name: "my-project"}
				if got_args.project != want_project_params {
					t.Errorf("got project params %v, want %v", got_args.project, want_project_params)
				}
		
				got_member_user_id := got_args.member.UserID
				want_member_user_id := test_uuid(test_user_id)
				if got_member_user_id != want_member_user_id {
					t.Errorf("got member user id %v, want %v", got_member_user_id, want_member_user_id)
				}
		
				got_role := got_args.member.Role
				want_role := pgtype.Text{String: "owner", Valid: true}
				if got_role != want_role {
					t.Errorf("got member role %v, want %v", got_role, want_role)
				}
			})
		}
	})

	cases := []struct {
		name     string
		repo_err error
	}{
		{"should return a repository error unchanged", errors.New("commit failed")},
		{"should return a duplicate key error unchanged", errors.New(`ERROR: duplicate key value violates unique constraint "projects_name_key" (SQLSTATE 23505)`)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()
			defer user_service.Clear()
			defer org_service.Clear()

			user_service.find_by_id_return = &database.User{UserID: test_uuid(test_user_id)}
			org_service.find_by_id_and_role_return = &database.FindOneOrganizationByIdAndRoleRow{
				Role: "owner",
			}
			project_repository.create_project_with_member_err = c.repo_err

			got_project, err := project_service.Create(test_user_id, test_org_id, "my-project")

			if got_project != nil {
				t.Errorf("got project %v, want nil", got_project)
			}

			got_error := err
			want_error := c.repo_err
			if got_error != want_error {
				t.Errorf("got error %v, want %v", got_error, want_error)
			}
		})
	}
}

func test_uuid(id string) pgtype.UUID {
	uuid := pgtype.UUID{}
	uuid.Scan(id)
	return uuid
}

func TestProjectServiceDeleteOne(t *testing.T) {
	ctx := context.Background()
	project_repository := &StubProjectRepository{}
	project_service := NewService(ctx, slog.Default(), project_repository, &StubUserService{}, &StubOrgService{})

	t.Run("should reject the deletion with 'invalid_role' when the member is not an owner or editor", func(t *testing.T) {
		tests := []struct {
			role string
		}{
			{"viewer"},
			{"kambing"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("role: '%s'", tt.role), func(t *testing.T) {
				defer project_repository.Clear()

				project_repository.find_one_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
					ProjectID: test_uuid(test_project_id),
					Role:      pgtype.Text{String: tt.role, Valid: true},
				}

				got_err := project_service.DeleteOne(test_user_id, test_project_id)

				assert_error(t, got_err, errors.New("invalid_role"))

				got_find_args := project_repository.find_one_by_id_and_role_call_args
				want_find_args := database.FindOneProjectByIdAndRoleParams{ProjectID: test_uuid(test_project_id), UserID: test_uuid(test_user_id)}
				if len(got_find_args) != 1 {
					t.Fatalf("got %d FindOneByIdAndRole calls, want 1", len(got_find_args))
				}
				if got_find_args[0] != want_find_args {
					t.Errorf("got params %v, want %v", got_find_args[0], want_find_args)
				}

				got_n_calls := project_repository.delete_project_with_members_n_calls
				want_n_calls := 0
				if got_n_calls != want_n_calls {
					t.Errorf("got %d DeleteProjectWithMembers calls, want %d", got_n_calls, want_n_calls)
				}
			})
		}
	})

	lookup_cases := []struct {
		name     string
		repo_err error
	}{
		{"should reject the deletion with 'invalid_role' when the user is not a project member", pgx.ErrNoRows},
		{"should reject the deletion with 'invalid_role' when the project lookup fails", errors.New("db down")},
	}

	for _, c := range lookup_cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_one_by_id_and_role_return = nil
			project_repository.find_one_by_id_and_role_err = c.repo_err

			got_err := project_service.DeleteOne(test_user_id, test_project_id)

			assert_error(t, got_err, errors.New("invalid_role"))

			got_n_calls := project_repository.delete_project_with_members_n_calls
			want_n_calls := 0
			if got_n_calls != want_n_calls {
				t.Errorf("got %d DeleteProjectWithMembers calls, want %d", got_n_calls, want_n_calls)
			}
		})
	}

	cases := []struct {
		name     string
		role     string
		repo_err error
	}{
		{"should delete the project with its members as owner", "owner", nil},
		{"should delete the project with its members as editor", "editor", nil},
		{"should return the repository error", "owner", errors.New("db unreachable")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_one_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
				ProjectID: test_uuid(test_project_id),
				Role:      pgtype.Text{String: c.role, Valid: true},
			}
			project_repository.delete_project_with_members_err = c.repo_err

			err := project_service.DeleteOne(test_user_id, test_project_id)

			got_error := err
			want_error := c.repo_err
			if got_error != want_error {
				t.Errorf("got error %v, want %v", got_error, want_error)
			}

			got_n_calls := project_repository.delete_project_with_members_n_calls
			want_n_calls := 1
			if got_n_calls != want_n_calls {
				t.Fatalf("got %d DeleteProjectWithMembers calls, want %d", got_n_calls, want_n_calls)
			}

			got_project_id := project_repository.delete_project_with_members_call_args[0]
			want_project_id := test_uuid(test_project_id)
			if got_project_id != want_project_id {
				t.Errorf("got project id %v, want %v", got_project_id, want_project_id)
			}
		})
	}
}

func TestProjectServiceFindById(t *testing.T) {
	ctx := context.Background()
	project_repository := &StubProjectRepository{}
	project_service := NewService(ctx, slog.Default(), project_repository, &StubUserService{}, &StubOrgService{})

	found := &database.FindOneProjectByIdRow{ProjectID: test_uuid(test_project_id)}

	cases := []struct {
		name        string
		repo_return *database.FindOneProjectByIdRow
		repo_err    error
		want_return *database.FindOneProjectByIdRow
		want_err    error
	}{
		{"should return nil without error when there are no rows", nil, pgx.ErrNoRows, nil, nil},
		{"should return a generic error when the query fails", nil, errors.New("db down"), nil, errors.New("unable to find project")},
		{"should return the project when found", found, nil, found, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_one_by_id_return = c.repo_return
			project_repository.find_one_by_id_err = c.repo_err

			got_return, got_err := project_service.FindById(test_user_id, test_project_id)

			assert_error(t, got_err, c.want_err)
			if got_return != c.want_return {
				t.Errorf("got project %v, want %v", got_return, c.want_return)
			}

			got_args := project_repository.find_one_by_id_call_args
			want_args := database.FindOneProjectByIdParams{ProjectID: test_uuid(test_project_id), UserID: test_uuid(test_user_id)}
			if len(got_args) != 1 {
				t.Fatalf("got %d FindOneById calls, want 1", len(got_args))
			}
			if got_args[0] != want_args {
				t.Errorf("got params %v, want %v", got_args[0], want_args)
			}
		})
	}
}

func TestProjectServiceFindByIdAndRole(t *testing.T) {
	ctx := context.Background()
	project_repository := &StubProjectRepository{}
	project_service := NewService(ctx, slog.Default(), project_repository, &StubUserService{}, &StubOrgService{})

	found := &database.FindOneProjectByIdAndRoleRow{ProjectID: test_uuid(test_project_id)}

	cases := []struct {
		name        string
		repo_return *database.FindOneProjectByIdAndRoleRow
		repo_err    error
		want_return *database.FindOneProjectByIdAndRoleRow
		want_err    error
	}{
		{"should return nil without error when there are no rows", nil, pgx.ErrNoRows, nil, nil},
		{"should return a generic error when the query fails", nil, errors.New("db down"), nil, errors.New("unable to find user")},
		{"should return the project when found", found, nil, found, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_one_by_id_and_role_return = c.repo_return
			project_repository.find_one_by_id_and_role_err = c.repo_err

			got_return, got_err := project_service.FindByIdAndRole(test_user_id, test_project_id, []string{"owner"})

			assert_error(t, got_err, c.want_err)
			if got_return != c.want_return {
				t.Errorf("got project %v, want %v", got_return, c.want_return)
			}

			got_args := project_repository.find_one_by_id_and_role_call_args
			want_args := database.FindOneProjectByIdAndRoleParams{ProjectID: test_uuid(test_project_id), UserID: test_uuid(test_user_id)}
			if len(got_args) != 1 {
				t.Fatalf("got %d FindOneByIdAndRole calls, want 1", len(got_args))
			}
			if got_args[0] != want_args {
				t.Errorf("got params %v, want %v", got_args[0], want_args)
			}
		})
	}
}

func TestProjectServiceListMyProjects(t *testing.T) {
	ctx := context.Background()
	project_repository := &StubProjectRepository{}
	project_service := NewService(ctx, slog.Default(), project_repository, &StubUserService{}, &StubOrgService{})

	rows := []database.FindProjectsForUserRow{
		{ProjectID: test_uuid(test_project_id)},
	}

	cases := []struct {
		name        string
		repo_return []database.FindProjectsForUserRow
		repo_err    error
		want_return []database.FindProjectsForUserRow
		want_err    error
	}{
		{"should return an empty list when there are no rows", nil, pgx.ErrNoRows, []database.FindProjectsForUserRow{}, nil},
		{"should return an error when the query fails", nil, errors.New("db down"), nil, errors.New("unable to find project users, db query failed")},
		{"should return the user's projects", rows, nil, rows, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_for_user_return = c.repo_return
			project_repository.find_for_user_err = c.repo_err

			got_return, got_err := project_service.ListMyProjects(test_user_id)

			assert_error(t, got_err, c.want_err)

			if c.want_return == nil {
				if got_return != nil {
					t.Errorf("got projects %v, want nil", got_return)
				}
			}
			if c.want_return != nil {
				if got_return == nil {
					t.Fatalf("got projects nil, want %v", c.want_return)
				}
				if len(got_return) != len(c.want_return) {
					t.Fatalf("got %d projects, want %d", len(got_return), len(c.want_return))
				}
				for i := range c.want_return {
					if got_return[i] != c.want_return[i] {
						t.Errorf("got project %v, want %v", got_return[i], c.want_return[i])
					}
				}
			}

			got_args := project_repository.find_for_user_call_args
			want_user_id := test_uuid(test_user_id)
			if len(got_args) != 1 {
				t.Fatalf("got %d FindForUser calls, want 1", len(got_args))
			}
			if got_args[0] != want_user_id {
				t.Errorf("got user id %v, want %v", got_args[0], want_user_id)
			}
		})
	}
}

func TestProjectServiceUpdateOne(t *testing.T) {
	ctx := context.Background()
	project_repository := &StubProjectRepository{}
	project_service := NewService(ctx, slog.Default(), project_repository, &StubUserService{}, &StubOrgService{})

	t.Run("should reject the update with 'invalid_role' when the member is not an owner or editor", func(t *testing.T) {
		tests := []struct {
			role string
		}{
			{"viewer"},
			{"kambing"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("role: '%s'", tt.role), func(t *testing.T) {
				defer project_repository.Clear()

				project_repository.find_one_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
					ProjectID: test_uuid(test_project_id),
					Role:      pgtype.Text{String: tt.role, Valid: true},
				}

				got_return, got_err := project_service.UpdateOne(test_user_id, test_project_id, "renamed")

				assert_error(t, got_err, errors.New("invalid_role"))
				if got_return != nil {
					t.Errorf("got project %v, want nil", got_return)
				}

				got_find_args := project_repository.find_one_by_id_and_role_call_args
				want_find_args := database.FindOneProjectByIdAndRoleParams{ProjectID: test_uuid(test_project_id), UserID: test_uuid(test_user_id)}
				if len(got_find_args) != 1 {
					t.Fatalf("got %d FindOneByIdAndRole calls, want 1", len(got_find_args))
				}
				if got_find_args[0] != want_find_args {
					t.Errorf("got params %v, want %v", got_find_args[0], want_find_args)
				}

				got_n_calls := project_repository.update_one_n_calls
				want_n_calls := 0
				if got_n_calls != want_n_calls {
					t.Errorf("got %d UpdateOne calls, want %d", got_n_calls, want_n_calls)
				}
			})
		}
	})

	lookup_cases := []struct {
		name     string
		repo_err error
	}{
		{"should reject the update with 'invalid_role' when the user is not a project member", pgx.ErrNoRows},
		{"should reject the update with 'invalid_role' when the project lookup fails", errors.New("db down")},
	}

	for _, c := range lookup_cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_one_by_id_and_role_return = nil
			project_repository.find_one_by_id_and_role_err = c.repo_err

			got_return, got_err := project_service.UpdateOne(test_user_id, test_project_id, "renamed")

			assert_error(t, got_err, errors.New("invalid_role"))
			if got_return != nil {
				t.Errorf("got project %v, want nil", got_return)
			}

			got_n_calls := project_repository.update_one_n_calls
			want_n_calls := 0
			if got_n_calls != want_n_calls {
				t.Errorf("got %d UpdateOne calls, want %d", got_n_calls, want_n_calls)
			}
		})
	}

	updated := &database.Project{ProjectID: test_uuid(test_project_id), Name: "renamed"}

	cases := []struct {
		name        string
		role        string
		repo_return *database.Project
		repo_err    error
		want_return *database.Project
	}{
		{"should send the name and project id and stamp updated_at as owner", "owner", updated, nil, updated},
		{"should send the name and project id and stamp updated_at as editor", "editor", updated, nil, updated},
		{"should return the repository error", "owner", nil, errors.New("db down"), nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer project_repository.Clear()

			project_repository.find_one_by_id_and_role_return = &database.FindOneProjectByIdAndRoleRow{
				ProjectID: test_uuid(test_project_id),
				Role:      pgtype.Text{String: c.role, Valid: true},
			}
			project_repository.update_one_return = c.repo_return
			project_repository.update_one_err = c.repo_err

			before := time.Now()
			got_return, got_err := project_service.UpdateOne(test_user_id, test_project_id, "renamed")

			if got_err != c.repo_err {
				t.Errorf("got error %v, want %v", got_err, c.repo_err)
			}
			if got_return != c.want_return {
				t.Errorf("got project %v, want %v", got_return, c.want_return)
			}

			got_args := project_repository.update_one_call_args
			if len(got_args) != 1 {
				t.Fatalf("got %d UpdateOne calls, want 1", len(got_args))
			}
			if got_args[0].Name != "renamed" {
				t.Errorf("got name %q, want %q", got_args[0].Name, "renamed")
			}
			if got_args[0].ProjectID != test_uuid(test_project_id) {
				t.Errorf("got project id %v, want %v", got_args[0].ProjectID, test_uuid(test_project_id))
			}
			if !got_args[0].UpdatedAt.Valid {
				t.Errorf("got invalid UpdatedAt, want a valid timestamp")
			}
			if got_args[0].UpdatedAt.Time.Before(before) {
				t.Errorf("got UpdatedAt %v, want at or after %v", got_args[0].UpdatedAt.Time, before)
			}
		})
	}
}

func assert_error(t *testing.T, got error, want error) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("got error %v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Errorf("got error nil, want %v", want)
		return
	}
	if got.Error() != want.Error() {
		t.Errorf("got error %v, want %v", got, want)
	}
}
