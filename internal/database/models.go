// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Password  string
	Username  string
	IsPremium bool
}
