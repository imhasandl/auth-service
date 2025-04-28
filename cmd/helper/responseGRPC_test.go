package helper

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRespondWithErrorGRPC(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		code     codes.Code
		message  string
		err      error
		expected string
	}{
		{
			name:     "Internal error underlying error",
			code:     codes.Internal,
			message:  "database error",
			err:      errors.New("connection failed"),
			expected: "database error",
		},
		{
			name:     "Not found error without underlying error",
			code:     codes.NotFound,
			message:  "user not found",
			err:      nil,
			expected: "user not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RespondWithErrorGRPC(ctx, tc.code, tc.message, tc.err)

			// Verify the error is a status error
			statusErr, ok := status.FromError(err)
			assert.True(t, ok, "Error should be a gRPC status error")

			// Verify the code matches
			assert.Equal(t, tc.code, statusErr.Code())

			// Verify the message contains our expected text
			assert.True(t, strings.Contains(statusErr.Message(), tc.expected))
		})
	}
}
