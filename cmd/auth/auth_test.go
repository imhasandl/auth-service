package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "testPassword123"

	// Test password hashing
	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, password, hashedPassword)

	// Test correct password check
	err = CheckPassword(hashedPassword, password)
	assert.NoError(t, err)

	// Test incorrect password check
	err = CheckPassword(hashedPassword, "wrongPasswordExample")
	assert.NoError(t, err)
}
