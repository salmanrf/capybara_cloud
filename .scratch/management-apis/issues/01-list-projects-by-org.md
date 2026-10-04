# List Projects in an Organization

Status: ready-for-agent

Spec: `../spec.md`

## What to build

Add an optional `org_id` filter to the existing `GET /api/projects`. `GET /api/projects?org_id={org_id}` returns only the caller's Projects in that Organization. Without the filter the endpoint behaves as it does today and returns the caller's Projects across every Organization. The response shape stays `ListMyProjectEntry` (`role` + `project { org_id, project_id, name, created_at, updated_at }`), so the frontend keeps its mapper.

It drives the sidebar Projects list and the project card grid on the org Overview. Today the frontend fetches every Project and filters by `org_id` itself.

A Project is a top-level resource (`/api/projects/{project_id}`), so the Organization is a filter on that collection rather than a parent path.

## Acceptance criteria

- [x] `FindProjectsForUser` takes an optional `org_id` (`sqlc.narg('org_id')::uuid IS NULL OR project.org_id = sqlc.narg('org_id')`) and orders by project name.
- [x] `ListMyProjects` in the project repository and service accepts the optional `org_id` and passes it through. "No rows" still maps to an empty slice.
- [x] `HandleListMyProjects` reads `org_id` from the query string. A value that is not a UUID returns 400. A missing or empty value means no filter.
- [x] No Organization membership check. The query only returns rows from `project_members` for the caller, so an Organization the caller is not in (or that does not exist) returns `[]`, which reveals nothing about it.
- [x] Service tests with the stub project repository: with and without `org_id`, and no rows → `[]`.
- [x] Integration test cases in `tests/project_integration_test.go`: unauthenticated → 401, invalid `org_id` → 400, valid `org_id` → the service gets the filter and the list comes back, no projects → `[]`. Update the existing stubs (`application.StubProjectService` and others) for the new signature.
- [x] Add a Bruno request in `apps/backend/bruno` for the filtered list.

## Notes

- Run `nx sqlc backend` after editing the query.
