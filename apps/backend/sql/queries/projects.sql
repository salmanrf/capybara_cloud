-- name: CreateProject :one
INSERT INTO "projects" (org_id, name) VALUES ($1, $2) RETURNING *;

-- name: CreateProjectMember :one
INSERT INTO "project_members" 
  (project_id, user_id, role) 
VALUES 
  ($1, $2, $3) 
RETURNING *;

-- name: FindProjectsForUser :many
SELECT 
  "project_members".project_id, "project_members".user_id, "project_members".role,
  "project_members".created_at project_members_created_at,
  "project".org_id, "project".name, "project".created_at project_created_at, "project".updated_at project_updated_at
FROM "project_members" AS project_members
LEFT JOIN 
  "projects" AS project ON "project_members".project_id = "project".project_id
WHERE
  "project_members".user_id = $1;

-- name: FindOneProjectById :one
SELECT "project".*
FROM 
  "project_members" AS "project_members"
LEFT JOIN "projects" AS "project" ON "project_members".project_id = "project".project_id 
WHERE 
  "project".project_id = $1 AND "project_members".user_id = $2;

-- name: FindOneProjectByIdAndRole :one
SELECT "project".*, "project_members".role "role"
FROM 
  "project_members" AS "project_members"
LEFT JOIN "projects" AS "project" ON "project_members".project_id = "project".project_id 
WHERE 
  "project".project_id = $1 AND "project_members".user_id = $2;

-- name: UpdateOneProject :one
UPDATE "projects" 
SET 
  name = $1, 
  updated_at = $2 
WHERE 
  project_id = $3
RETURNING *;

-- name: DeleteOneProject :exec
DELETE FROM "projects" WHERE project_id = $1;

-- name: DeleteProjectMembersByProjectId :exec
DELETE FROM "project_members" WHERE project_id = $1;
