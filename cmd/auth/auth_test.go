package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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
	assert.Error(t, err)
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "token-secret"
	expiresAt := time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Checks that token has expected format
	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts))
}

func TestMakeRefreshToken(t *testing.T) {
	token1, err := MakeRefreshToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	// Verify a different token is generated each time
	token2, err := MakeRefreshToken()
	assert.NoError(t, err)
	assert.NotEqual(t, token1, token2)

	// Verify length (64 chars for 32 bytes hex encoded)
	assert.Equal(t, 64, len(token1))
}

func TestGenerateVerificationCode(t *testing.T) {
	code, err := GenerateVerificationCode()
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify code is in range 1000-9999
	assert.GreaterOrEqual(t, code, int32(1000))
	assert.LessOrEqual(t, code, int32(9999))

	// Verify different codes are generated
	code2, err := GenerateVerificationCode()

	// Checks if the second verification code is same as first
	if code == code2 {
		code2, err = GenerateVerificationCode()
		assert.NoError(t, err)
		assert.NotEqual(t, code, code2)
	}

	assert.NoError(t, err)
}