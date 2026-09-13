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

-- name: FindOneApplicationComplete :one
SELECT
  "app".*,
  sqlc.embed(config),
  sqlc.embed(pm),
  sqlc.embed(project),
  sqlc.embed(org)
FROM
  "applications" AS "app"
INNER JOIN
  "projects" AS "project"
    ON
      "project".project_id = "app".project_id
INNER JOIN
  "organizations" AS "org"
    ON
      "org".org_id = "project".org_id
LEFT JOIN
  "project_members" AS "pm"
    ON
      "pm".project_id = "app".project_id
      AND
      "pm".user_id = $2
LEFT JOIN
  "application_configs" AS "config"
    ON
      "config".app_id = "app".app_id
WHERE
  "app".app_id = $1
LIMIT 1;

-- name: CreateApplicationConfig :one
INSERT INTO "application_configs" (
  app_id,
  port,
  variables_json
)
VALUES ($1, $2, $3) 
ON CONFLICT (app_id)
DO UPDATE SET 
  port = $2,
  variables_json = $3, 
  updated_at = NOW()
RETURNING *;