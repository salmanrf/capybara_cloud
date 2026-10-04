# Management APIs for Masbro Dashboard v2

Status: needs-triage

Design source: Claude Design project `2b05f82e-f1ff-442a-9213-77179cb82ffd`, file `Masbro Dashboard v2.dc.html`. It has three screens: the org Overview (`view: 'overview'`), the project Applications canvas (`tab: 'apps'`) and the project Members tab (`tab: 'members'`).

## Problem Statement

The backend only offers basic CRUD for single resources. Organizations, Projects and Applications can be created, read, updated and deleted. Configs can be read and set, and a bundle can be uploaded to start a Deployment. The v2 dashboard needs collections scoped to a parent (the Projects of an Organization, the Applications of a Project, the members of either) and status or summary data that no endpoint returns. Several things the design shows are not modelled at all.

## What exists today

| Area | Routes |
| --- | --- |
| Auth | `GET /api/auth/me`, `POST /api/auth/signup`, `POST /api/auth/signin` |
| Organizations | `POST /`, `GET /` (mine), `GET/PUT/DELETE /{org_id}` |
| Projects | `POST /`, `GET /` (mine, across every Organization), `GET/PUT/DELETE /{project_id}` |
| Applications | `POST /`, `GET/PUT /{app_id}`, `GET/POST /{app_id}/configs`, `POST /{app_id}/deployments` |

## Gaps, by screen

| Design element | Gap | Issue |
| --- | --- | --- |
| Overview: sidebar Projects list, project cards | No per-org project listing | 01 |
| Overview: "Organization · N members", Members nav | No org member listing | 02 |
| Canvas: app cards with status | No per-project application listing, no derived status | 03 |
| Overview: "deployed 12m ago" / "failed 1h ago" | No deployment history | 04 |
| Members tab: table | No project member listing | 05 |
| Members tab: role dropdown, ⋯ menu, Invite Member | No member mutation; role vocabulary mismatch | 06 |
| Overview: app chips + last deploy on each card | Aggregation strategy undecided | 07 |
| Canvas: access links, `API_URL` edge labels, link panel | Not modelled; needs deploy-time injection | 08 |
| Overview badge "production" / header "production ▾" | Environments not modelled | 09 |
| Sidebar footer "node-01 · 3 of 8 GB used" | No hosts / capacity reporting | 10 |
| Git connections nav, "3 repos", commit SHA on app cards | No git integration | 11 |
| App card domain ("api.masbro.app ↗") and runtime ("node 20") | Not modelled | 12 |
| Draggable canvas cards | Positions not persisted | 13 |

Issues 01–05 need only new queries, handlers and routes on existing tables, so they are `ready-for-agent`. The rest need a decision first and are `needs-triage`.

## Shared implementation decisions (01–05)

- Each endpoint follows the existing layering: a sqlc query in `apps/backend/sql/queries`, a repository method, a `Service` method, a handler in `api/handlers`, a route in `api/routes` behind `middleware.LoginGuard`, and a response DTO in `pkg/dto`.
- Access: a caller who is not a member of the parent resource gets the same not-found response that `GET /{id}` gives for that resource today. An empty collection returns `[]`, not an error.
- Responses use the existing envelope (`packages/shared-go/utils/http_response.go`) and snake_case JSON, matching `ListMyOrgEntry` / `ListMyProjectEntry`.
- Collections are unpaginated for now and ordered by name (or by `version_number DESC` for deployments).
- The code follows `.claude/rules/code-conventions.md`: docstrings, block control flow, no multi-line conditions, snake_case locals.
- Tests: service tests with the hand-written stubs in each package's `stubs.go`, plus an HTTP case per route in `apps/backend/tests/*_integration_test.go` (route mounted, auth required, not-found for non-members, happy path shape).

## Known gaps found while scoping

- `FindByIdAndRole` on both the project and organization services ignores its `roles` argument. The queries `FindOneProjectByIdAndRole` and `FindOneOrganizationByIdAndRole` select the role but never filter on it. Every role check passes for any member today. Fixing it is a prerequisite for 06.
- Project access is decided only by a `project_members` row. The Members tab says "Org owners always have access", which isn't true today. This is tracked in 06.

## Out of Scope

- Frontend wiring. The org Overview spec (`.scratch/organization-overview`) uses fixtures; swapping them for these endpoints is separate work.
- Search / ⌘K, Docs, Feedback, the avatar menu.
- Pagination and filtering.
