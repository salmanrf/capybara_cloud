# List Organization members

Status: ready-for-agent

Spec: `../spec.md`

## What to build

`GET /api/organizations/{org_id}/members` returns each member's `user_id`, `full_name`, `email`, `role` and `joined_at` (`organization_users.created_at`).

The org card's "Organization · N members" and the Members nav count use the length of this list. There is no org Members page in the design yet, so only listing is in scope.

## Acceptance criteria

- [ ] New sqlc query `FindOrganizationMembers(org_id)`: `organization_users` joined to `users`, ordered by `full_name`. Never select `hashed_password`.
- [ ] Organization service method plus handler and route. A caller who is not in the Organization gets the same not-found as `GET /api/organizations/{org_id}`.
- [ ] Response DTO in `pkg/dto/organization.go`.
- [ ] Service tests and an HTTP integration test case (401, non-member not-found, list).
- [ ] Bruno request.

## Notes

- The organization service still talks to `Queries` directly (no repository). Either add the query call there, matching the current style, or do the organization repository refactor first, as the project-repository spec suggests. Don't do both in one change.
- Inviting, changing roles and removing org members are out of scope until the design draws an org Members page.
