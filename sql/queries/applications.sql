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
SELECT "app".*, sqlc.embed(config), "pm".project_id pm_project_id, "pm".role role
FROM 
  "applications" AS "app"
LEFT JOIN 
  "project_members" as "pm" 
    ON 
      "pm".project_id = "app".project_id
      AND
      "pm".user_id = $2
LEFT JOIN
  "application_configs" as "config"
    ON
      "config".app_id = "app".app_id
WHERE 
  "app".app_id = $1
LIMIT 1;

-- name: CreateApplicationConfig :one
INSERT INTO "application_configs" (
  app_id,
  variables_json
)
VALUES ($1, $2) 
ON CONFLICT (app_id)
DO UPDATE SET variables_json = $2, updated_at = NOW()
RETURNING *;