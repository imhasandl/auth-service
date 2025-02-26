package helper

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RespondWithErrorGRPC(ctx context.Context, code codes.Code, msg string, err error) error {
	if err != nil {
		log.Println(err)
	}
	
	if code > codes.Internal { // 5XX equivalent in gRPC
		log.Printf("Responding with 5XX gRPC error: %s", msg)
	}

	type errorResponse struct {
		AuthServiceError string `json:"error"`
	}

	jsonBytes, err := json.Marshal(errorResponse{AuthServiceError: msg})
	if err != nil {
		log.Printf("Error marshalling error JSON: %s", err)
		return status.Errorf(codes.Internal, "Failed to marshal error response")
	}

	log.Printf("AuthServiceError: %s, Code: %s", string(jsonBytes), code.String()) // Log the error
	return status.Errorf(code, msg)
}