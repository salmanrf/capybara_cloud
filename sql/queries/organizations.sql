-- name: CreateOrganization :one
INSERT INTO "organizations" (name) VALUES ($1) RETURNING *;

-- name: CreateOrganizationUser :one
INSERT INTO "organization_users" (org_id, user_id, role) VALUES ($1, $2, $3) RETURNING *;

-- name: FindOrganizationsForUser :many
SELECT 
  "orgus".org_id, "orgus".user_id, "orgus".role,
  "orgus".created_at orgus_created_at,
  "org".name, "org".created_at org_created_at, "org".updated_at org_updated_at
FROM "organization_users" AS orgus
LEFT JOIN 
  "organizations" AS org ON "orgus".org_id = "org".org_id
WHERE
  "orgus".user_id = $1;