package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-service/cmd/auth"
	"github.com/imhasandl/auth-service/internal/database"
	"github.com/imhasandl/auth-service/internal/mocks"
	pb "github.com/imhasandl/auth-service/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegister(t *testing.T) {
	testCases := []struct {
		name          string
		request       *pb.RegisterRequest
		mockSetup     func(*mocks.MockQueries)
		expectedError bool
		errorCode     codes.Code
		errorMsg      string
	}{
		{
			name: "succesful registration",
			request: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testusername",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("CreateUser", mock.Anything, mock.MatchedBy(func(arg database.CreateUserParams) bool {
					return arg.Email == "test@example.com" && arg.Username == "testuser"
				})).Return(database.User{
					ID:               uuid.New(),
					Email:            "test@example.com",
					Username:         "testuser",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					IsPremium:        false,
					VerificationCode: 1234,
					IsVerified:       false,
				})

				mockDB.On("StoreVerificationCode", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "username too short",
			request: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "test", // Less than 5 characters
			},
			mockSetup:     func(mockDB *mocks.MockQueries) {},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "username should be at 5 characters or above",
		},
		{
			name: "database error during user creation",
			request: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testuser",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("CreateUser", mock.Anything, mock.Anything).Return(database.User{}, errors.New("database error"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't create user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			tc.mockSetup(mockDB)

			response, err := server.Register(ctx, tc.request)

			if tc.expectedError {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.errorCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tc.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tc.request.Email, response.User.Email)
				assert.Equal(t, tc.request.Username, response.User.Username)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	testCases := []struct {
		name          string
		request       *pb.VerifyEmailRequest
		mockSetup     func(mockDB *mocks.MockQueries)
		expectedError bool
		errorCode     codes.Code
		errorMsg      string
	}{
		{
			name: "successfull verification",
			request: &pb.VerifyEmailRequest{
				Email:            "test@example.com",
				VerificationCode: 1234,
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, database.GetUserByIdentifierParams{
					Email:    "test@example.com",
					Username: "",
				}).Return(database.User{
					ID:                     uuid.New(),
					Email:                  "test@example.com",
					Username:               "testuser",
					VerificationCode:       1234,
					IsVerified:             false,
					VerificationExpireTime: time.Now().Add(time.Hour), // Not expired
				}, nil)

				mockDB.On("VerifyUser", mock.Anything, "test@example.com").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "user not found",
			request: &pb.VerifyEmailRequest{
				Email:            "notfound@example.com",
				VerificationCode: 1234,
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{}, errors.New("user not found"))
			},
			expectedError: true,
			errorCode:     codes.NotFound,
			errorMsg:      "can't get user with identifier",
		},
		{
			name: "already verified",
			request: &pb.VerifyEmailRequest{
				Email:            "notfound@example.com",
				VerificationCode: 1234,
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{
					IsVerified: true,
				}, nil)
			},
			expectedError: true,
			errorCode:     codes.AlreadyExists,
			errorMsg:      "user is already verified",
		},
		{
			name: "invalid verification code",
			request: &pb.VerifyEmailRequest{
				Email:            "test@example.com",
				VerificationCode: 5678, // Different code than what's in the database
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{
					VerificationCode:       1234, // Actual code in DB
					IsVerified:             false,
					VerificationExpireTime: time.Now().Add(-time.Hour), // Expired 1 hour ago
				})
			},
			expectedError: true,
			errorCode:     codes.Unauthenticated,
			errorMsg:      "invalid verification code",
		},
		{
			name: "verification code expired",
			request: &pb.VerifyEmailRequest{
				Email:            "test@example.com",
				VerificationCode: 1234,
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{
					VerificationCode:       1234,
					IsVerified:             false,
					VerificationExpireTime: time.Now().Add(-time.Hour), // Expired 1 hour ago
				}, nil)
			},
			expectedError: true,
			errorCode:     codes.DeadlineExceeded,
			errorMsg:      "verification code expired",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			tc.mockSetup(mockDB)

			response, err := server.VerifyEmail(ctx, tc.request)

			if tc.expectedError {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.errorCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tc.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Success)
				assert.Equal(t, "Email verified successfully", response.Message)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name          string
		request       *pb.LoginRequest
		mockSetup     func(*mocks.MockQueries)
		expectedError bool
		errorCode     codes.Code
		errorMsg      string
	}{
		{
			name: "successfully logged in",
			request: &pb.LoginRequest{
				Identifier: "test@example.com",
				Password:   "password123",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				userID := uuid.New()

				// Hash the password for the mock user
				hashedPassword, err := auth.HashPassword("password123")
				assert.NoError(t, err)

				expectedParams := database.GetUserByIdentifierParams{
					Email:    "test@example.com",
					Username: "test@example.com",
				}
				mockDB.On("GetUserByIdentifier", mock.Anything, expectedParams).Return(database.User{
					ID:       userID,
					Email:    "test@example.com",
					Username: "testuser",
					Password: hashedPassword, // Use hashed password
				}, nil)

				mockDB.On("RefreshToken", mock.Anything, mock.MatchedBy(func(arg database.RefreshTokenParams) bool {
					return arg.UserID == userID
				})).Return(database.RefreshToken{
					Token:      "test-refresh-token",
					UserID:     userID,
					ExpiryTime: time.Now().Add(time.Hour * 7 * 24),
					CreatedAt:  time.Now(),
				}, nil)
			},
			expectedError: false,
			errorCode:     codes.OK,
			errorMsg:      "",
		},
		{
			name: "user not found",
			request: &pb.LoginRequest{
				Identifier: "test@example.com",
				Password:   "password123",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{}, errors.New("User not found"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't get user with identifier - Login",
		},
		{
			name: "database error storing refresh token",
			request: &pb.LoginRequest{
				Identifier: "test@example.com",
				Password:   "password123",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				userID := uuid.New()
				hashedPassword, err := auth.HashPassword("password123")
				assert.NoError(t, err)

				expectedUserParams := database.GetUserByIdentifierParams{
					Email:    "test@example.com",
					Username: "test@example.com",
				}
				mockDB.On("GetUserByIdentifier", mock.Anything, expectedUserParams).Return(database.User{
					ID:       userID,
					Password: hashedPassword,
				}, nil)

				mockDB.On("RefreshToken", mock.Anything, mock.Anything).Return(database.RefreshToken{}, errors.New("database error"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't store refresh token - Login",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			tc.mockSetup(mockDB)

			response, err := server.Login(ctx, tc.request)

			if tc.expectedError {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.errorCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tc.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.User)
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.RefreshToken)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestLogout(t *testing.T) {
	testCases := []struct {
		name          string
		request       *pb.LogoutRequest
		mockSetup     func(*mocks.MockQueries)
		expectedError bool
		errorCode     codes.Code
		errorMsg      string
	}{
		{
			name: "successful logout",
			request: &pb.LogoutRequest{
				RefreshToken: "test-logout",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("DeleteRefreshTokenByToken", mock.Anything, "test-logout").Return(nil)
			},
			expectedError: false,
			errorCode:     codes.OK,
			errorMsg:      "successfully logged out",
		},
		{
			name: "database error during logout",
			request: &pb.LogoutRequest{
				RefreshToken: "test-logout",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("DeleteRefreshTokenByToken", mock.Anything, "test-logout").Return(errors.New("database error"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't delete token - Logout",
		},
		{
			name: "empty refresh token",
			request: &pb.LogoutRequest{
				RefreshToken: "",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				// No DB call expected with empty token
			},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
			errorMsg:      "refresh token is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			tc.mockSetup(mockDB)

			response, err := server.Logout(ctx, tc.request)

			if tc.expectedError {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.errorCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tc.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Success)
				assert.Equal(t, "successfully logged out", tc.errorMsg)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	testCases := []struct {
		name          string
		request       *pb.RefreshTokenRequest
		mockSetup     func(*mocks.MockQueries)
		expectedError bool
		errorCode     codes.Code
		errorMsg      string
	}{
		{
			name: "successfuly refreshed token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "test-refresh-token",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				userID := uuid.New()
				mockDB.On("GetRefreshToken", mock.Anything, "test-refresh-token").Return(database.RefreshToken{
					Token:      "test-refresh-token",
					UserID:     userID,
					ExpiryTime: time.Now().Add(time.Hour * 7 * 24),
					CreatedAt:  time.Now(),
				}, nil)

				mockDB.On("DeleteTokenByUserID", mock.Anything, userID).Return(nil)
				mockDB.On("RefreshToken", mock.Anything, mock.Anything).Return(database.RefreshToken{
					Token:      "new-refresh-token",
					UserID:     userID,
					ExpiryTime: time.Now().Add(time.Hour * 7 * 24),
					CreatedAt:  time.Now(),
				}, nil)
			},
			expectedError: false,
			errorCode:     codes.OK,
			errorMsg:      "",
		},
		{
			name: "can't get token from database",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "wrong-token",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetRefreshToken", mock.Anything, "wrong-token").Return(database.RefreshToken{}, errors.New("token not found"))
			},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
			errorMsg:      "can't get refresh token - RefreshToken",
		},
		{
			name: "empty refresh token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				// No DB call with empty body request
			},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
			errorMsg:      "refresh token is required",
		},
		{
			name: "expired refresh token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "expired-token",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				userID := uuid.New()
				mockDB.On("GetRefreshToken", mock.Anything, "expired-token").Return(database.RefreshToken{
					Token:      "expired-token",
					UserID:     userID,
					ExpiryTime: time.Now().Add(-time.Hour),
					CreatedAt:  time.Now().Add(-time.Hour * 7 * 24),
				}, nil)
			},
			expectedError: true,
			errorCode:     codes.DeadlineExceeded,
			errorMsg:      "refresh token expired - RefreshToken",
		},
		{
			name: "error deleting old token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid-token",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				userID := uuid.New()
				mockDB.On("GetRefreshToken", mock.Anything, "valid-token").Return(database.RefreshToken{
					Token:      "valid-token",
					UserID:     userID,
					ExpiryTime: time.Now().Add(time.Hour * 7 * 24),
					CreatedAt:  time.Now(),
				}, nil)

				mockDB.On("DeleteTokenByUserID", mock.Anything, userID).Return(errors.New("database error"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't delete old token - RefreshToken",
		},
		{
			name: "error storing new refresh token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid-token",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				userID := uuid.New()
				mockDB.On("GetRefreshToken", mock.Anything, "valid-token").Return(database.RefreshToken{
					Token:      "valid-token",
					UserID:     userID,
					ExpiryTime: time.Now().Add(time.Hour * 7 * 24),
					CreatedAt:  time.Now(),
				}, nil)

				mockDB.On("DeleteTokenByUserID", mock.Anything, userID).Return(nil)
				mockDB.On("RefreshToken", mock.Anything, mock.Anything).Return(database.RefreshToken{}, errors.New("database error"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't store refresh token - RefreshToken",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			tc.mockSetup(mockDB)

			response, err := server.RefreshToken(ctx, tc.request)

			if tc.expectedError {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.errorCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tc.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NotNil(t, response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestSendVerificationCode(t *testing.T) {
	testCases := []struct {
		name          string
		request       *pb.SendVerifyCodeRequest
		mockSetup     func(*mocks.MockQueries)
		expectedError bool
		errorCode     codes.Code
		errorMsg      string
	}{
		{
			name: "successfully send verification code",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
