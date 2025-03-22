// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const createUser = `-- name: CreateUser :one
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
RETURNING id, created_at, updated_at, email, password, username, subscribers, subscribed_to, is_premium, verification_code, is_verified
`

type CreateUserParams struct {
	ID               uuid.UUID
	Email            string
	Password         string
	Username         string
	IsPremium        bool
	VerificationCode int32
	IsVerified       bool
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Email,
		arg.Password,
		arg.Username,
		arg.IsPremium,
		arg.VerificationCode,
		arg.IsVerified,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.Password,
		&i.Username,
		pq.Array(&i.Subscribers),
		pq.Array(&i.SubscribedTo),
		&i.IsPremium,
		&i.VerificationCode,
		&i.IsVerified,
	)
	return i, err
}

const getUserByIdentifier = `-- name: GetUserByIdentifier :one
SELECT id, created_at, updated_at, email, password, username, subscribers, subscribed_to, is_premium, verification_code, is_verified FROM users
WHERE email = $1 OR username = $2
`

type GetUserByIdentifierParams struct {
	Email    string
	Username string
}

func (q *Queries) GetUserByIdentifier(ctx context.Context, arg GetUserByIdentifierParams) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByIdentifier, arg.Email, arg.Username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.Password,
		&i.Username,
		pq.Array(&i.Subscribers),
		pq.Array(&i.SubscribedTo),
		&i.IsPremium,
		&i.VerificationCode,
		&i.IsVerified,
	)
	return i, err
}

const sendVerifyCodeAgain = `-- name: SendVerifyCodeAgain :exec
UPDATE users 
SET verification_code = $1, is_verified = FALSE
WHERE id = $2
`

type SendVerifyCodeAgainParams struct {
	VerificationCode int32
	ID               uuid.UUID
}

func (q *Queries) SendVerifyCodeAgain(ctx context.Context, arg SendVerifyCodeAgainParams) error {
	_, err := q.db.ExecContext(ctx, sendVerifyCodeAgain, arg.VerificationCode, arg.ID)
	return err
}

const storeVerificationCode = `-- name: StoreVerificationCode :exec
UPDATE users 
SET verification_code = $1, is_verified = FALSE
WHERE id = $2
`

type StoreVerificationCodeParams struct {
	VerificationCode int32
	ID               uuid.UUID
}

func (q *Queries) StoreVerificationCode(ctx context.Context, arg StoreVerificationCodeParams) error {
	_, err := q.db.ExecContext(ctx, storeVerificationCode, arg.VerificationCode, arg.ID)
	return err
}

const verifyUser = `-- name: VerifyUser :exec
UPDATE users 
SET is_verified = TRUE, verification_code = 0
WHERE email = $1
`

func (q *Queries) VerifyUser(ctx context.Context, email string) error {
	_, err := q.db.ExecContext(ctx, verifyUser, email)
	return err
}
