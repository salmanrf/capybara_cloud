-- name: CreateApplication :one
INSERT INTO "applications" (
  project_id,
  type,
  name
) 
VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateOneApplication :one
UPDATE "applications"
SET 
  name = $2,
  updated_at = $3
WHERE 
  app_id = $1
RETURNING *;

-- name: FindOneApplicationWithProjectMember :one
SELECT "app".*, "pm".role "role"
FROM 
  "applications" AS "app"
LEFT JOIN 
  "project_members" as "pm" 
    ON 
      "pm".project_id = "app".project_id
      AND
      "pm".user_id = $2
WHERE 
  "app".app_id = $1 AND "pm".role IS NOT NULL
LIMIT 1;