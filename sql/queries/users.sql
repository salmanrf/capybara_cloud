-- name: GetUsersByEmail :one
SELECT * FROM "users" WHERE "email" = $1 LIMIT 1;

-- name: CreateOneUser :one
INSERT INTO "users" (
  username,
  email,
  full_name,
  role
) VALUES (
$1, $2, $3, $4
) RETURNING *;