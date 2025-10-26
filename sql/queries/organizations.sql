-- name: CreateOrganization :one
INSERT INTO "organizations" (name) VALUES ($1) RETURNING *;

-- name: CreateOrganizationUser :one
INSERT INTO "organization_users" (org_id, user_id, role) VALUES ($1, $2, $3) RETURNING *;