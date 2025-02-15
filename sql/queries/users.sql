-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password, username, is_premium)
VALUES (
   $1,
   NOW(),
   NOW(),
   $2,
   $3,
   $4,
   $5
)
RETURNING *;

-- name: GetUserByIdentifier :one
SELECT * FROM users
WHERE email = $1 OR username = $2;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetAllUsers :many
SELECT * FROM users;
