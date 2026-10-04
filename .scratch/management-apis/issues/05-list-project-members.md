# List Project members

Status: ready-for-agent

Spec: `../spec.md`

## What to build

`GET /api/projects/{project_id}/members` returns each member's `user_id`, `full_name`, `email`, `role` and `joined_at` (`project_members.created_at`), ordered by `full_name`.

It drives the table on the project Members tab and the "Members 3" count in the project sidebar. The UI marks the current user "(you)" by comparing with `/auth/me`, so the backend does not need to.

## Acceptance criteria

- [ ] New sqlc query `FindProjectMembers(project_id)`: `project_members` joined to `users`. Never select `hashed_password`.
- [ ] Project repository + service method; handler + route on the project router. A caller who is not a project member gets the same not-found as `GET /api/projects/{project_id}`.
- [ ] Response DTO in `pkg/dto/project.go`.
- [ ] Service tests (stub repository) and an HTTP integration test case.
- [ ] Bruno request.

## Out of scope

- Changing roles, removing members and inviting (06).
- Showing org owners who aren't project members (06 decides whether they appear).
