# Manage Project members (role change, remove, invite)

Status: needs-triage
Blocked by: 05

Spec: `../spec.md`

## What the design needs

On the project Members tab: a role dropdown per row (Owner / Developer / Viewer), a ⋯ menu (presumably Remove), and an "Invite Member" primary action. The page copy says "Org owners always have access."

Likely routes: `POST /api/projects/{project_id}/members`, `PATCH /api/projects/{project_id}/members/{user_id}`, `DELETE /api/projects/{project_id}/members/{user_id}`.

## Decisions needed

1. **Role vocabulary.** The backend writes `owner`, and the application service checks for `member`. The design uses Owner / Developer / Viewer. What is the canonical set, and what can each role do (deploy? edit configs? manage members?). Existing `member` rows need a migration.
2. **Role enforcement.** `FindOneProjectByIdAndRole` and `FindOneOrganizationByIdAndRole` ignore their `roles` argument, so every member passes every role check today. This has to be fixed before role-gated mutations mean anything.
3. **Invite flow.** Can you only add people who already have an account (by email lookup), or does inviting create a pending invitation with a token and email? The second needs a new table and a mailer.
4. **Org owner access.** Should org owners get implicit access to every Project (a change in the access query), or should they be added as explicit `project_members`? Should they appear in the members list?
5. **Last owner guard.** Can the last Owner be demoted or removed?
