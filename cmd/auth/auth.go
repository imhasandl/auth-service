package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TokenType represents the type of authentication token
type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "media-access"
)

// MakeJWT generates a JWT token for the specified user ID
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

// MakeRefreshToken generates a secure random token for refresh authentication
func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// HashPassword hashes the user's password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}

// MockCheckPassword is a mock function for password checking used in tests
var MockCheckPassword = func(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// CheckPassword checks if the provided password matches the hashed password
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateVerificationCode generates a random 4-digit verification code
func GenerateVerificationCode() (int32, error) {
	var code int32
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 0, err
	}
	code = 1000 + (int32(b[0])|int32(b[1])<<8|int32(b[2])<<16|int32(b[3])<<24)%9000
	return code, nil
}

// SendVerificationEmail sends an email with a verification code using SMTP protocol
func SendVerificationEmail(email, emailSender, emailSecret string, code int32) error {
	from := emailSender
	password := emailSecret
	to := email
	subject := "Email Verification"
	body := fmt.Sprintf("Your verification code is: %d", code)
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
