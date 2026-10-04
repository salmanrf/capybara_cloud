# Spec: List Projects in an Organization

Status: needs-triage

Parent spec: `spec.md` (Management APIs for Masbro Dashboard v2). Issue: `issues/01-list-projects-by-org.md`.

## Problem Statement

The v2 dashboard is scoped to one Organization at a time. On the org Overview, the sidebar Projects list and the project card grid should only show that Organization's Projects. The backend can only list every Project the caller belongs to, across every Organization. So the frontend downloads the whole list on every Overview load and filters by `org_id` itself. That wastes payload as a user joins more Organizations, puts scoping logic in the client, and leaves the order of the list undefined, so the sidebar can reorder between loads.

## Solution

The existing "my Projects" collection, `GET /api/projects`, gets an optional `org_id` query filter. `GET /api/projects?org_id={org_id}` returns only the caller's Projects in that Organization, ordered by Project name. Without the filter the endpoint returns the caller's Projects across all Organizations, as it does today, but now also ordered by name. The response shape does not change, so the frontend keeps its existing mapper and only adds the query parameter.

A Project is a top-level resource (`/api/projects/{project_id}`), so the Organization is a filter on that collection rather than a parent path such as `/api/organizations/{org_id}/projects`.

## User Stories

1. As a dashboard user, I want the sidebar to list only the Projects of the Organization I'm viewing, so that I'm not distracted by Projects from other Organizations.
2. As a dashboard user, I want the org Overview card grid to show only that Organization's Projects, so that the Overview reflects the Organization I picked.
3. As a dashboard user, I want Projects listed alphabetically by name, so that I can find one quickly and the list doesn't reorder between page loads.
4. As a dashboard user who belongs to many Organizations, I want the Overview to load only the Projects it shows, so that it stays fast as my memberships grow.
5. As a dashboard user, I want to see my role on each Project in the list, so that I know what I can do before I open it.
6. As a dashboard user, I want to see only the Projects I'm a member of in an Organization, not every Project in it, so that I don't see Projects I can't open.
7. As a dashboard user with no Projects in an Organization, I want an empty list rather than an error, so that the Overview can show an empty state and a "create Project" prompt.
8. As a dashboard user who switches Organizations, I want each switch to fetch that Organization's Projects, so that the sidebar always matches the selected Organization.
9. As a frontend developer, I want the filtered list to use the same entry shape as the unfiltered list, so that I can reuse the existing mapper and types.
10. As a frontend developer, I want to keep calling the same endpoint and just add a query parameter, so that the change to the API client is one argument.
11. As a frontend developer, I want the unfiltered `GET /api/projects` to keep working, so that screens that need every Project, such as a future cross-org switcher, are not broken.
12. As a frontend developer, I want a malformed `org_id` to return a 400 with a clear message, so that client bugs surface as validation errors and not as 500s.
13. As a frontend developer, I want an empty `org_id` (`?org_id=`) to behave like no filter, so that an unset value in the client doesn't cause an error.
14. As an unauthenticated visitor, I want to get a 401 from the endpoint, so that Project data is never exposed without a session.
15. As a user who is not in an Organization, I want filtering by that Organization to return an empty list, so that the API doesn't reveal whether the Organization exists or what is in it.
16. As a security reviewer, I want the filtered list scoped by Project membership, so that knowing an `org_id` grants no access to that Organization's Projects.
17. As a security reviewer, I want a nonexistent `org_id` and an `org_id` the caller isn't in to return the same response, so that the endpoint can't be used to probe which Organizations exist.
18. As a backend developer, I want one query to serve both the filtered and the unfiltered list, so that the two can't drift apart in shape or ordering.
19. As a backend developer, I want the filter to go through the existing project service method rather than a new one, so that the Project domain keeps one way to list a user's Projects.
20. As a backend developer, I want the project service to stay independent of the organization service for this feature, so that no new cross-domain dependency is wired in `main.go`.
21. As a QA engineer, I want a Bruno request for the filtered list, so that I can smoke-test it against a running backend.
22. As a maintainer, I want integration tests for 401, 400, filtered, unfiltered and empty cases, so that regressions in routing and validation are caught without a database.

## Implementation Decisions

- **API contract.** `GET /api/projects` accepts an optional `org_id` query parameter.
  - Missing or empty: no filter. Returns every Project the caller is a member of.
  - A valid UUID: returns only the caller's Projects whose Project `org_id` matches.
  - Anything else: 400 with a validation message, and the project service is not called.
  - Success is 200 with the standard response envelope. The data is a list of `ListMyProjectEntry` (`role` plus `project { org_id, project_id, name, created_at, updated_at }`). An empty result is `[]`, never `null`.
  - The route stays behind the login guard. A missing or invalid `sid` returns 401.
- **No Organization membership check.** Rows come from the caller's `project_members` rows, so the filter can only narrow what the caller can already see. An Organization the caller isn't in, or one that doesn't exist, returns `[]`. That is the normal response for a filtered collection and reveals nothing. This replaces the earlier draft's "same not-found as `GET /api/organizations/{org_id}`" requirement and the plan to inject `organization.Service` into the project service.
- **Query.** The existing sqlc query `FindProjectsForUser` gains a nullable `org_id` parameter (a `sqlc.narg`, matched when null or equal) and an `ORDER BY` on Project name. One query serves both forms. No new query and no schema change.
- **Repository.** `ProjectRepository.FindForUser` takes the caller's user UUID plus the optional Organization UUID (or the generated params struct) and passes both to the query.
- **Service.** `project.Service.ListMyProjects` takes `user_id` and an optional `org_id` (a nullable value, empty meaning no filter). It still maps "no rows" to an empty slice and any other error to the generic "db query failed" error.
- **Handler.** `HandleListMyProjects` reads `org_id` from the URL query, validates it as a UUID when present, and passes it to the service. The DTO mapping (`NewListMyProjectResponse`) does not change.
- **Ordering.** Projects are ordered by name for both the filtered and the unfiltered list. This changes the unfiltered response from an undefined order to a defined one, which is backward compatible.
- **Follow-on (issue 07).** Options B and C of the project card summary now refer to `GET /projects?org_id=…` and `GET /projects/summary?org_id=…`.

## Testing Decisions

- **Good tests check external behavior:** HTTP status codes, response bodies, service return values and errors, and which arguments reached the next layer. They don't check how a function computes its result internally.
- **Seam 1, HTTP (integration).** Mount the project router over the stub project service and drive it with `httptest`. Add a table-driven `TestProjectListMine` to the project integration test file, following the existing project integration tests. Cases:
  - No `sid` → 401, and the service is not called.
  - `org_id=not-a-uuid` → 400, and the service is not called.
  - No `org_id` → 200 with the stubbed list, and the service is called once with no filter.
  - A valid `org_id` → 200 with the stubbed list, and the service is called once with that `org_id`.
  - Service returns an empty slice → 200 with `[]`.
  - Service returns an error → 500.

  The stub project service records the `org_id` it receives so the test can assert on it.
- **Seam 2, project service (domain logic).** Build the real service with `NewService` over the stub project repository. Extend the existing `TestProjectServiceListMyProjects`:
  - With and without `org_id`, the right UUIDs reach the repository (recorded call args).
  - A "no rows" error → empty slice and no error.
  - Any other repository error → the generic error.
- **SQL behavior is not unit-tested.** The `org_id` filter and the name ordering live in the query, and there is no database in either test layer. They are covered by the Bruno smoke request against a local Supabase.
- **Stubs.** Update every `ListMyProjects` / `FindForUser` stub for the new signature: the backend test stubs, the application package's project service stub and the project package's repository stub. Add call-arg recording rather than adding a mocking library.
- **Prior art:** the project integration tests (`TestProjectGetOne`, `TestProjectUpdateOne`) and the reference application deployment integration test for the HTTP seam, and the existing project service tests for the domain seam.

## Out of Scope

- `GET /api/organizations/{org_id}/projects` or any other nested route for Projects.
- Distinguishing "Organization not found" from "not a member" for this endpoint.
- Listing every Project in an Organization regardless of Project membership, for example for Organization owners or admins.
- Pagination, search and sort parameters other than the fixed name ordering.
- Embedding Applications, statuses or last-deployment data in each entry. That is issue 07.
- Frontend changes beyond passing `org_id`. Those belong to the frontend work for the v2 dashboard.

## Further Notes

- Run `nx sqlc backend` after editing the query, because the generated database package is gitignored.
- The `organization_users` and `project_members` tables are independent. A user can in principle hold a Project membership in an Organization they have left. This endpoint returns such Projects, the same as the unfiltered list does today. Closing that gap belongs with the member-management work (issue 06), not here.
- Issue 01 has already been rewritten to match this spec.
