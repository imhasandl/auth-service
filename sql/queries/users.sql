-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password, username, is_premium, verification_code, is_verified)
VALUES (
   $1,
   NOW(),
   NOW(),
   $2,
   $3,
   $4,
   $5,
   $6,
   $7
)
RETURNING *;

-- name: GetUserByIdentifier :one
SELECT * FROM users
WHERE email = $1 OR username = $2;

-- name: StoreVerificationCode :exec
UPDATE users 
SET verification_code = $1, is_verified = false 
WHERE id = $2;

-- name: VerifyUser :exec
UPDATE users 
SET is_verified = true, verification_code = "no-code"
WHERE email = $1;