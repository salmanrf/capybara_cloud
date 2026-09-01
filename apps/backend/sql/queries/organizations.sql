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

-- name: FindOneOrganizationById :one
SELECT "org".*
FROM 
  "organization_users" AS "orgus"
LEFT JOIN "organizations" AS "org" ON "orgus".org_id = "org".org_id 
WHERE 
  "org".org_id = $1 AND "orgus".user_id = $2;

-- name: FindOneOrganizationByIdAndRole :one
SELECT "org".*, "orgus".role "role"
FROM 
  "organization_users" AS "orgus"
LEFT JOIN "organizations" AS "org" ON "orgus".org_id = "org".org_id 
WHERE 
  "org".org_id = $1 AND "orgus".user_id = $2;

-- name: UpdateOneOrganization :one
UPDATE "organizations" 
SET 
  name = $1, 
  updated_at = $2 
WHERE 
  org_id = $3
RETURNING *;

-- name: DeleteOneOrganization :exec
DELETE FROM "organizations" WHERE org_id = $1;

-- name: DeleteOrganizationUsersByOrgId :exec
DELETE FROM "organization_users" WHERE org_id = $1;

-- -- name: FindOneOrganizationByIdAndRole :one
-- SELECT "org".*, "orgus".role "role"
-- FROM 
--   "organization_users" AS "orgus"
-- LEFT JOIN "organizations" AS "org" ON "orgus".org_id = "org".org_id 
-- WHERE 
--   "org".org_id = $1 AND "orgus".user_id = $2 AND "orgus".role = ANY(@roles::varchar[]);