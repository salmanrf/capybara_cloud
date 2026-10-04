# List an Application's Deployments

Status: ready-for-agent

Spec: `../spec.md`

## What to build

`GET /api/applications/{app_id}/deployments` returns the Application's Deployments, newest first (`version_number DESC`). Each entry has `app_dp_id`, `version_number`, `status` (the raw int), a derived `outcome` (`building | succeeded | failed`), `created_at` and `updated_at`.

It backs the "deployed 12m ago" / "failed 1h ago" footer on project cards and any future deploy history view.

## Acceptance criteria

- [ ] New sqlc query `FindDeploymentsByAppId(app_id)`. Do not return `artifacts_path`, `build_path` or `variables_snapshot_json`. Those are internal and could hold secrets.
- [ ] A deployment repository and service method. Access goes through the existing application access check (the caller must be a member of the app's Project), as `POST /{app_id}/deployments` does.
- [ ] `outcome`: `status < 0` → `failed`, `5` → `succeeded`, otherwise `building`. Unit-test it.
- [ ] Service tests and an HTTP integration test case on the existing application router.
- [ ] Bruno request.

## Notes

- Add an optional `?limit=` only if a caller needs it; the project card only needs the latest deployment, and 07 decides how that is fetched.
