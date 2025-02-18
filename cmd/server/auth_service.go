package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/imhasandl/auth-service/cmd/helper/auth"
	"github.com/imhasandl/auth-service/internal/database"
	pb "github.com/imhasandl/auth-service/protos"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	db          *database.Queries
	tokenSecret string
}

func NewServer(db *database.Queries, tokenSecret string) *server {
	return &server{
		pb.UnimplementedAuthServiceServer{},
		db,
		tokenSecret,
	}
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(req.GetUsername()) < 5 {
		return nil, status.Errorf(codes.Internal, "username should be 5 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
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

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword()))
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