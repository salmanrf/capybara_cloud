# List Applications in a Project, with status

Status: ready-for-agent

Spec: `../list-project-applications-spec.md`

## What to build

`GET /api/applications?project_id={project_id}` returns every Application in the Project, ordered by name. An Application is a top-level resource (`/api/applications/{app_id}`), so the Project is a query filter on that collection, the same pattern as `GET /api/projects?org_id=` (issue 01).

Each entry has:

- `app_id`, `name`, `type`, `created_at`, `updated_at`
- `status`: `not_deployed | running | stopped`, derived from the Application's current Deployment Instance
- `latest_deployment`: `{ app_dp_id, version_number, status, outcome, created_at, updated_at }` or `null`

This drives the cards on the Applications canvas, and the app chips with status dots on the Overview project cards (see 07).

## Derivation

`status` comes only from the Application's newest `deployment_instances` row (by `created_at`), whichever Deployment it belongs to:

| Current Deployment Instance | `status` |
| --- | --- |
| none | `not_deployed` |
| `DEPLOY_INSTANCE_STATUS_RUNNING` | `running` |
| `STOPPED` or any other value | `stopped` |

`latest_deployment.outcome` comes from the Deployment with the highest `version_number`. This is the same rule as issue 04, written once in the deployment domain:

| Deployment step | `outcome` |
| --- | --- |
| < 0 | `failed` |
| 1–4 | `building` |
| 5 | `succeeded` |
| anything else | `failed` |

A rollout that is building or has failed never changes `status`. The previous container keeps showing `running`.

## Acceptance criteria

- [ ] New sqlc query: the Project's Applications, inner-joined to `project_members` on the caller, with one `LEFT JOIN LATERAL` for the newest instance and another for the latest Deployment. Check the `sqlc.yaml` nullable overrides for the left-joined columns.
- [ ] Application repository method, and `application.Service.ListByProject(user_id, project_id)` returning response entries with `status` and `outcome` derived. "No rows" → `[]`.
- [ ] No separate access check. A non-member or a nonexistent Project → `[]`.
- [ ] A handler and route `GET /` on the application router. A missing or non-UUID `project_id` → 400, and the service is not called.
- [ ] Entries never include the artifact path, build path, variables snapshot, container name or host port.
- [ ] Service tests (stub repository): UUIDs reach the repository; no rows → `[]`; repository error → the generic error; a table case per status row and per outcome row; instance running while the latest Deployment is building, and while it has failed → `running`.
- [ ] HTTP integration test `TestListApplications`: 401, missing `project_id` → 400, invalid `project_id` → 400, filter reaches the service, mixed statuses with a `null` `latest_deployment`, empty `[]`, service error → 500.
- [ ] Bruno request.

## Out of scope

- Instance status accuracy. Until issue 14 lands, every deployed Application shows `stopped`.
- Domain, runtime label and commit SHA on the card (12, 11).
- Canvas positions (13) and access links (08).
