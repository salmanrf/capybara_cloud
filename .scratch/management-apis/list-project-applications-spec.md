# Spec: List Applications in a Project, with status

Status: needs-triage

Parent spec: `spec.md` (Management APIs for Masbro Dashboard v2). Issue: `issues/03-list-project-applications.md`. Sibling decision: `list-projects-by-org-spec.md`.

## Problem Statement

The project Applications canvas in the v2 dashboard draws one card per Application, each with a status dot. The Overview's project cards show the same Applications as chips with status dots (issue 07). The backend can't list a Project's Applications: only `GET /api/applications/{app_id}` exists, so the frontend would need every `app_id` in advance. The backend also doesn't say whether an Application is actually up. That depends on its Deployment Instance (the container started by the Start step), and the frontend can't see that data.

## Solution

Add a filtered collection, `GET /api/applications?project_id={project_id}`. It returns every Application in the Project. Each entry has:

- a `status` derived from the Application's current Deployment Instance, which says whether it is running;
- a summary of the latest Deployment, with its `outcome`, which says whether a rollout is building or has failed.

An Application is a top-level resource: it is created with `POST /api/applications` and read with `GET /api/applications/{app_id}`. So the Project is a query filter on that collection, which is the same pattern `GET /api/projects?org_id=` uses.

`project_id` is required. If the caller isn't a member of the Project, or the Project doesn't exist, the response is an empty list.

## User Stories

1. As a dashboard user, I want the Applications canvas to show every Application in the Project I'm viewing, so that I can see the whole Project at a glance.
2. As a dashboard user, I want each Application card to show whether its container is running, so that I know it is serving traffic.
3. As a dashboard user, I want an Application whose container is stopped to show as stopped, so that I know it is down even though it was deployed.
4. As a dashboard user, I want an Application that has never had a container to show as not deployed, so that I know it needs a first deploy.
5. As a dashboard user, I want an Application that is still running its previous container while a new Deployment fails to keep showing as running, so that a failed rollout doesn't make me think the live app is down.
6. As a dashboard user, I want an Application that is still running its previous container while a new Deployment builds to keep showing as running, so that I can see it is up while I wait for the new version.
7. As a dashboard user, I want to see separately that the latest Deployment is building, so that I know a rollout is in progress and I shouldn't start another.
8. As a dashboard user, I want to see separately that the latest Deployment failed, so that I can investigate and redeploy even while the old version keeps serving.
9. As a dashboard user, I want each card to show the latest Deployment's version number and when it last changed, so that I know how fresh the code is.
10. As a dashboard user, I want each card to show the Application's name and type, so that I can recognise it.
11. As a dashboard user, I want the canvas to load in one request, so that it renders quickly even in a Project with many Applications.
12. As a dashboard user in a Project with no Applications, I want an empty list rather than an error, so that the canvas can show an empty state with a "create Application" prompt.
13. As a frontend developer, I want `status` to be one of `not_deployed`, `running` or `stopped`, so that I can map it straight to a dot colour and label.
14. As a frontend developer, I want `latest_deployment.outcome` to be one of `building`, `succeeded` or `failed`, so that I can show rollout progress without interpreting raw step numbers.
15. As a frontend developer, I want `latest_deployment` to be `null` when there is none, so that I can branch on it without checking magic values.
16. As a frontend developer, I want the raw Deployment step next to the derived outcome, so that I can show finer progress later without an API change.
17. As a frontend developer, I want listing Applications to follow the same filter pattern as listing Projects, so that the API client and my mental model stay consistent.
18. As a frontend developer, I want a missing or malformed `project_id` to return 400, so that client bugs show up as validation errors and not as 500s or silently empty lists.
19. As a frontend developer, I want status and outcome values defined in one place on the backend, so that the canvas chips, the Overview chips (issue 07) and the deploy history (issue 04) never disagree.
20. As an unauthenticated visitor, I want to get a 401, so that Application data is never exposed without a session.
21. As a user who is not a member of a Project, I want filtering by that Project to return an empty list, so that the API doesn't reveal whether the Project exists or what it contains.
22. As a security reviewer, I want the list scoped by the caller's Project membership in the query itself, so that knowing a `project_id` grants nothing.
23. As a security reviewer, I want a nonexistent `project_id` and a `project_id` the caller isn't a member of to return the same response, so that the endpoint can't be used to probe which Projects exist.
24. As a security reviewer, I want the list to leave out Deployment and instance internals (artifact path, build path, variables snapshot, container name, host port), so that secrets and server details don't leak through the dashboard.
25. As a backend developer, I want the Applications, their current Deployment Instance and their latest Deployment fetched in one query, so that the endpoint doesn't do N+1 queries per Application.
26. As a QA engineer, I want a Bruno request for the filtered list, so that I can smoke-test status and outcome against real Deployments.
27. As a maintainer, I want every status and outcome rule covered by a table-driven test, so that a change to the instance or deploy step constants can't silently change what the dashboard shows.

## Implementation Decisions

- **API contract.** `GET /api/applications` on the existing application router, behind the login guard.
  - `project_id` query parameter is required. Missing or empty → 400. Not a UUID → 400. In both cases the application service is not called.
  - Success is 200 with the standard response envelope. The data is a list of entries ordered by Application name: `{ app_id, name, type, created_at, updated_at, status, latest_deployment }`.
  - `latest_deployment` is `{ app_dp_id, version_number, status, outcome, created_at, updated_at }` (`status` is the raw Deployment step int), or `null`.
  - Entries never include the Deployment's artifact path, build path or variables snapshot, or the instance's container name or host port.
  - An empty result is `[]`, never `null`. A missing or invalid `sid` → 401.
- **`status` comes from the Deployment Instance only.** The current Deployment Instance is the Application's newest `deployment_instances` row (by `created_at`), whichever Deployment it belongs to. Each Start step creates a new instance and never updates older ones, so only the newest row is current.

  | Current Deployment Instance | `status` |
  | --- | --- |
  | none | `not_deployed` |
  | `DEPLOY_INSTANCE_STATUS_RUNNING` | `running` |
  | `DEPLOY_INSTANCE_STATUS_STOPPED` | `stopped` |
  | any other value | `stopped` |

  The latest Deployment never affects `status`. A rollout that is building or has failed leaves `status` showing whatever the current instance is. An unknown instance value maps to `stopped`, so it never shows as healthy.

  Today the Start step never writes `RUNNING` (see Known limitations), so until issue 14 lands, every deployed Application shows `stopped`.
- **`latest_deployment.outcome` comes from the Deployment.** The latest Deployment is the one with the highest `version_number`, the same rule `FindCurrentDeployment` uses.

  | Deployment step | `outcome` |
  | --- | --- |
  | < 0 | `failed` |
  | 1–4 (`INITIATED` … `BUILD_IMAGE_PUSHED`) | `building` |
  | 5 (`BUILD_INSTANCE_STARTED`) | `succeeded` |
  | anything else | `failed` |

  This is the same `outcome` issue 04 defines for deploy history. It is one function in the deployment domain, used by both issues.
- **Access is scoped in the query, with no separate check.** The query inner-joins `project_members` on the caller and the filtered `project_id`. A non-member or a nonexistent Project yields `[]`. The service does not call `project.Service` for this method.
- **No unfiltered list.** No screen needs every Application across every Project. Making the filter optional later is backward compatible.
- **Query.** A new sqlc query lists the filtered Project's Applications with two independent `LEFT JOIN LATERAL`s:
  - the Application's newest Deployment Instance (by `created_at`, keyed on `app_id`, not on the latest Deployment);
  - the Application's latest Deployment (by `version_number`).

  Columns from both left joins must come out as nullable `pgtype`s, so check the `sqlc.yaml` overrides for any NOT NULL columns that sqlc would otherwise type as non-nullable.
- **Repository.** The application repository gets a list method that takes the generated params (user UUID, project UUID) and returns the generated rows.
- **Service.** `application.Service` gets `ListByProject(user_id, project_id)`. It returns the response entries with `status` and `outcome` already derived, the way `FindOneConfig` already returns a DTO. "No rows" maps to an empty slice, and any other error to a generic query-failed error.
- **Derivation functions.** Both are pure functions over nullable inputs, and their result strings are exported constants:
  - Application `status` from the nullable instance status, in the application package.
  - Deployment `outcome` from the Deployment step, in the deployment domain, shared with issue 04.
- **Handler.** A new list handler validates `project_id`, calls the service and returns the entries. Validation follows the 400 message style of `GET /api/projects?org_id=`.
- **Issue 07.** Option A ("01, then 03 for each Project") now means one `GET /api/applications?project_id=` per Project. Option B embeds the same `status` and `latest_deployment.outcome`.

## Testing Decisions

- **What a good test checks:** HTTP status codes and bodies at the HTTP seam, and service return values, errors and what reached the repository stub at the service seam. Tests don't check how the query or the mapping is built.
- **Seam 1, HTTP (integration).** Mount the application router over the stub application service and drive it with `httptest`. Add a table-driven `TestListApplications` to the applications integration test file. Cases:
  - No `sid` → 401, and the service is not called.
  - Missing `project_id` → 400, and the service is not called.
  - `project_id=not-a-uuid` → 400, and the service is not called.
  - A valid `project_id` → 200, and the service is called once with the caller's user ID and that `project_id`.
  - The service returns entries with mixed statuses and outcomes, including a `null` `latest_deployment` → the body carries them through.
  - The service returns an empty slice → `[]`.
  - The service returns an error → 500.

  The stub application service gets a return value, an error, a call counter and recorded call args for the new method.
- **Seam 2, application service (domain logic).** Build the real service with `NewService` over the stub application repository. Cases:
  - User and Project UUIDs reach the repository.
  - "No rows" → `[]`.
  - Any other repository error → the generic error.
  - Table-driven status cases, one per row of the status table: no instance, running, stopped, unknown value.
  - Table-driven outcome cases, one per row of the outcome table: failed, each building step, succeeded, unknown step, plus no Deployment → `null`.
  - The cases that motivated this design: the current instance is running while the latest Deployment is building, and while it has failed. The expected result is `status: running` with the matching outcome in both.

  Each case feeds a repository row in and asserts on the entry's `status` and `latest_deployment` coming out. Testing the derivations through the service seam keeps the seam count at two. When issue 04 lands, its own tests cover `outcome` at its service seam.
- **SQL behavior is not unit-tested.** Membership scoping, picking the newest instance and the latest Deployment, and name ordering live in the query, and neither test layer has a database. The Bruno request against a local Supabase with seeded Deployments and instances covers them.
- **Prior art:** `TestFindOneApplication` / `TestCreateApplication` (HTTP seam) and `TestApplicationService` (service seam). The `TestProjectListMine` work from issue 01 is the closest template for the filter, 400 and empty-list cases.

## Out of Scope

- `GET /api/projects/{project_id}/applications` or any other nested route for Applications.
- An unfiltered `GET /api/applications` across every Project.
- Telling a nonexistent Project apart from one the caller isn't a member of.
- Reconciling instance status with the real container state. `status` reflects the `deployment_instances` row, and if a container dies without the row being updated, the dashboard still shows `running`. This is accepted for now and is fixed by issue 14 (see Further Notes).
- Stopping or cleaning up older instances when a new one starts.
- Deployment history (issue 04), the domain, runtime label and commit SHA on the card (issues 12 and 11), canvas positions (issue 13) and access links (issue 08).
- Embedding Applications in the project list (issue 07, option B).
- Live status updates by polling or push. The frontend refetches.
- Pagination.

## Further Notes

- The rule now used across the management APIs: a resource with its own top-level path (Projects, Applications) gets its collection filtered by a query parameter. A resource you can only reach through its parent stays nested: Organization and Project members (issues 02 and 05) and an Application's Deployments (issue 04).
- The design only draws running, building and failed dots. With this split, the frontend picks the dot from `status`, and overlays a building or failed badge from `latest_deployment.outcome` when the rollout isn't `succeeded`. How `not_deployed` and `stopped` look is a frontend decision.
- **Known limitations, accepted, fixed by issue 14.** Both come from how the deploy pipeline maintains `deployment_instances`, not from this endpoint, so the fix belongs there and this endpoint's contract stays the same:
  - The Start step creates the instance as `STOPPED` and never writes `RUNNING`, because `masbro-worker`'s `Start` passes the status through unchanged. Until issue 14 lands, every deployed Application shows `stopped`.
  - Nothing updates an instance when its container dies, so once `RUNNING` is written, `status` can show `running` for a dead container.
