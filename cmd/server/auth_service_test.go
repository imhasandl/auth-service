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
			name: "succssful login",
			request: &pb.LoginRequest{
				Identifier: "test@example.com",
				Password:   "password123",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Password: "hashed_password", // In real test, use properly hashed password
				}, nil)
			},
			expectedError: false,
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
			errorMsg:      "can't get user with identifier",
		},
		{
			name: "database error storing refresh token",
			request: &pb.LoginRequest{
				Identifier: "test@example.com",
				Password:   "password123",
			},
			mockSetup: func(mockDB *mocks.MockQueries) {
				mockDB.On("GetUserByIdentifier", mock.Anything, mock.Anything).Return(database.User{
					ID:       uuid.New(),
					Password: "hashed_password",
				}, nil)

				mockDB.On("RefreshToken", mock.Anything, mock.Anything).Return(database.RefreshToken{}, errors.New("database error"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
			errorMsg:      "can't store refresh token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			auth.MockCheckPassword = func(hashedPassword, password string) error {
				if tc.name == "successful login" {
					return nil
				}
				return errors.New("invalid password")
			}

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
        name           string
        request        *pb.LogoutRequest
        mockSetup      func(*mocks.MockQueries)
        expectedErrorr bool
        errorCode      codes.Code
        errorMsg       string
    }{
        {
            name: "successful logout",
            request: &pb.LogoutRequest{
                RefreshToken: "test-refresh-token",
            },
            mockSetup: func(mockDB *mocks.MockQueries) {
                mockDB.On("DeleteRefreshTokenByToken", mock.Anything, "test-refresh-token").Return(nil)
            },
            expectedErrorr: false,
            errorCode:      codes.OK,
            errorMsg:       "successfully log out",
        },
        {
            name: "database error during logout",
            request: &pb.LogoutRequest{
                RefreshToken: "test-refresh-token",
            },
            mockSetup: func(mockDB *mocks.MockQueries) {
                mockDB.On("DeleteRefreshTokenByToken", mock.Anything, "test-refresh-token").Return(errors.New("database error"))
            },
            expectedErrorr: true,
            errorCode:      codes.Internal,
            errorMsg:       "can't delete refresh token",
        },
        {
            name: "empty refresh token",
            request: &pb.LogoutRequest{
                RefreshToken: "",
            },
            mockSetup: func(mockDB *mocks.MockQueries) {
                // No DB call expected with empty token
            },
            expectedErrorr: true,
            errorCode:      codes.InvalidArgument,
            errorMsg:       "refresh token is required",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockDB := new(mocks.MockQueries)
            server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
            ctx := context.Background()

            tc.mockSetup(mockDB)

            response, err := server.Logout(ctx, tc.request)

            if tc.expectedErrorr {
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
                assert.Equal(t, "successfully logged out", response.Message)
            }
            mockDB.AssertExpectations(t)
        })
    }
}