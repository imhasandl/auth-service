package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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

	// Execute each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockDB := new(mocks.MockQueries)
			server := NewServer(mockDB, "test-secret", "test@example.com", "email-secret")
			ctx := context.Background()

			// Configure mocks based on the test case
			tc.mockSetup(mockDB)

			// Execute the test
			response, err := server.Register(ctx, tc.request)

			// Assert results
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
