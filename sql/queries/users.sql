-- name: FindOneUserByEmail :one
SELECT * FROM "users" WHERE "email" = $1 LIMIT 1;

-- name: FindOneUserByEmailOrId :one
SELECT * FROM "users" WHERE "user_id" = $1 LIMIT 1;

-- name: CreateOneUser :one
INSERT INTO "users" (
  username,
  email,
  full_name,
  hashed_password,
  role
) VALUES (
$1, $2, $3, $4, $5
) RETURNING *;