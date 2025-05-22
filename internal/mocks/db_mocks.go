package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-service/internal/database"
	"github.com/stretchr/testify/mock"
)

// MockQueries is a mock implementation of the database.Queries interface
type MockQueries struct {
	mock.Mock
}

// CreateUser mocks the CreateUser method
func (m *MockQueries) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.User), args.Error(1)
}

// GetUserByIdentifier mocks the GetUserByIdentifier method
func (m *MockQueries) GetUserByIdentifier(ctx context.Context, arg database.GetUserByIdentifierParams) (database.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.User), args.Error(1)
}

// VerifyUser mocks the VerifyUser method
func (m *MockQueries) VerifyUser(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

// StoreVerificationCode mocks the StoreVerificationCode method
func (m *MockQueries) StoreVerificationCode(ctx context.Context, arg database.StoreVerificationCodeParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// SendVerifyCodeAgain mocks the SendVerifyCodeAgain method
func (m *MockQueries) SendVerifyCodeAgain(ctx context.Context, arg database.SendVerifyCodeAgainParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// RefreshToken mocks the RefreshToken method
func (m *MockQueries) RefreshToken(ctx context.Context, arg database.RefreshTokenParams) (database.RefreshToken, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.RefreshToken), args.Error(1)
}

// GetRefreshToken mocks the GetRefreshToken method
func (m *MockQueries) GetRefreshToken(ctx context.Context, token string) (database.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(database.RefreshToken), args.Error(1)
}

// DeleteTokenByUserID mocks the DeleteTokenByUserID method
func (m *MockQueries) DeleteTokenByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// DeleteRefreshTokenByToken mocks the DeleteRefreshTokenByToken method
func (m *MockQueries) DeleteRefreshTokenByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
