// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: refresh_token.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const refreshToken = `-- name: RefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expiry_time) 
VALUES (
   $1, 
   $2, 
   $3
)
RETURNING token, user_id, expiry_time, created_at
`

type RefreshTokenParams struct {
	Token      string
	UserID     uuid.UUID
	ExpiryTime time.Time
}

func (q *Queries) RefreshToken(ctx context.Context, arg RefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, refreshToken, arg.Token, arg.UserID, arg.ExpiryTime)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.UserID,
		&i.ExpiryTime,
		&i.CreatedAt,
	)
	return i, err
}
