-- name: CreateApplication :one
INSERT INTO "applications" (
  project_id,
  type,
  name
) 
VALUES ($1, $2, $3) RETURNING *;