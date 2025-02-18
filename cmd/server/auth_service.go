package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-service/cmd/helper/auth"
	"github.com/imhasandl/auth-service/internal/database"
	pb "github.com/imhasandl/auth-service/protos"
	"github.com/imhasandl/post-service/cmd/helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	db          *database.Queries
	tokenSecret string
	emailSecret string
}

func NewServer(db *database.Queries, tokenSecret, emailSecret string) *server {
	return &server{
		pb.UnimplementedAuthServiceServer{},
		db,
		tokenSecret,
		emailSecret,
	}
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(req.GetUsername()) < 5 {
		return nil, status.Errorf(codes.Internal, "username should be 5 characters long")
	}

	hashedPassword, err := auth.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v - Register", err)
	}

	userParams := database.CreateUserParams{
		ID:       uuid.New(),
		Email:    req.GetEmail(),
		Password: string(hashedPassword),
		Username: req.GetUsername(),
	}

	user, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create user: %v - Register", err)
	}

	verificationCode, err := auth.verificationCode()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate verification code: %v - Register", err)
	}



	createdAtProto := timestamppb.New(user.CreatedAt) // Converts time.Time type into timespamppb
	updatedAtProto := timestamppb.New(user.UpdatedAt) // Converts time.Time type into timespamppb

	return &pb.RegisterResponse{
		User: &pb.User{
			Id:        user.ID.String(),
			CreatedAt: createdAtProto,
			UpdatedAt: updatedAtProto,
			Email:     user.Email,
			Username:  user.Username,
			IsPremium: user.IsPremium,
		},
	}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userParams := database.GetUserByIdentifierParams{
		Email:    req.GetIdentifier(),
		Username: req.GetIdentifier(),
	}

	user, err := s.db.GetUserByIdentifier(ctx, userParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get user with identifier: %v - Login", err)
	}

	createdAtProto := timestamppb.New(user.CreatedAt) // Converts time.Time type into timespamppb
	updatedAtProto := timestamppb.New(user.UpdatedAt) // Converts time.Time type into timespamppb

	err = auth.CheckPassword(user.Password, req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid credentials: %v - Login", err)
	}

	accessToken, err := auth.MakeJWT(user.ID, s.tokenSecret, time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create token: %v - Login", err)
	}

	return &pb.LoginResponse{
		User: &pb.User{
			Id:        user.ID.String(),
			CreatedAt: createdAtProto,
			UpdatedAt: updatedAtProto,
			Email:     user.Email,
			Username:  user.Username,
			IsPremium: user.IsPremium,
		},
		Token: accessToken,
	}, nil
}

func (s *server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	accessToken, err := helper.GetBearerTokenFromGrpc(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v - RefreshToken", err)
	}

	userID, err := helper.ValidateJWT(accessToken, s.tokenSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get token from header: %v - RefreshToken", err)
	}

	newAccessToken, err := auth.MakeJWT(userID, s.tokenSecret, time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create new access token: %v - RefreshToken", err)
	}

	refreshTokenParams := database.RefreshTokenParams{
		Token:      newAccessToken,
		UserID:     uuid.NullUUID{UUID: userID, Valid: true},
		ExpiryTime: time.Now().Add(24 * time.Hour),
	}

	refreshToken, err := s.db.RefreshToken(ctx, refreshTokenParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't store refresh token: %v - RefreshToken", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newAccessToken,
		ExpiryTime:   timestamppb.New(refreshToken.ExpiryTime),
	}, nil
}
