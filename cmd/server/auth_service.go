package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	auth "github.com/imhasandl/auth-service/cmd/helper"
	"github.com/imhasandl/auth-service/internal/database"
	pb "github.com/imhasandl/auth-service/protos"
	postService "github.com/imhasandl/post-service/cmd/helper"
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

	verificationCode := auth.GenerateVerificationCode()

	userParams := database.CreateUserParams{
		ID:               uuid.New(),
		Email:            req.GetEmail(),
		Password:         hashedPassword,
		Username:         req.GetUsername(),
		IsPremium:        false,
		VerificationCode: verificationCode,
		IsVerified:       false,
	}

	user, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create user: %v - Register", err)
	}

	verifyParams := database.StoreVerificationCodeParams{
		VerificationCode: verificationCode,
		ID:               user.ID,
	}

	err = s.db.StoreVerificationCode(ctx, verifyParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store verification code: %v - Register", err)
	}

	err = auth.SendVerificationEmail(req.GetEmail(), s.emailSecret, verificationCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send verification email: %v - Register", err)
	}

	return &pb.RegisterResponse{
		User: &pb.User{
			Id:               user.ID.String(),
			CreatedAt:        timestamppb.New(user.CreatedAt),
			UpdatedAt:        timestamppb.New(user.UpdatedAt),
			Email:            user.Email,
			Username:         user.Username,
			IsPremium:        user.IsPremium,
			VerificationCode: user.VerificationCode,
			IsVerified:       user.IsVerified,
		},
	}, nil
}

func (s *server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	userParams := database.GetUserByIdentifierParams{
		Email:    req.GetEmail(),
		Username: req.GetEmail(),
	}

	user, err := s.db.GetUserByIdentifier(ctx, userParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get user with identifier: %v - VerifyEmail", err)
	}

	if user.IsVerified {
		return nil, status.Errorf(codes.AlreadyExists, "email already verified: %v - VerifyEmail", err)
	}

	if user.VerificationCode != req.GetVerificationCode() {
		return nil, status.Errorf(codes.Unauthenticated, "invalid verification code: %v - VerifyEmail", err)
	}

	err = s.db.VerifyUser(ctx, req.GetEmail())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify user: %v - VerifyEmail", err)
	}

	return &pb.VerifyEmailResponse{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}

func (s *server) SendVerifyCodeAgain(ctx context.Context, req *pb.SendVerifyCodeAgainRequest) (*pb.SendVerifyCodeAgainResponse, error) {
	userParams := database.GetUserByIdentifierParams{
		Email:    req.GetEmail(),
		Username: req.GetEmail(),
	}
	
	user, err := s.db.GetUserByIdentifier(ctx, userParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get user with identifier: %v - SendVerifyCodeAgain", err)
	}

	newVerifyCode := auth.GenerateVerificationCode()

	sendVerifyAgainParams := database.SendVerifyCodeAgainParams{
		VerificationCode: newVerifyCode,
		ID: user.ID,
	}

	err = s.db.SendVerifyCodeAgain(ctx, sendVerifyAgainParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send verification code again: %v - SendVerifyCodeAgain", err)
	}

	err = auth.SendVerificationEmail(req.GetEmail(), s.emailSecret, newVerifyCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send verification email: %v - Register", err)
	}

	return &pb.SendVerifyCodeAgainResponse{
		Success: true,
		Message: "new verification code sent",
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
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
			Email:     user.Email,
			Username:  user.Username,
			IsPremium: user.IsPremium,
		},
		Token: accessToken,
	}, nil
}

func (s *server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	accessToken, err := postService.GetBearerTokenFromGrpc(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v - RefreshToken", err)
	}

	userID, err := postService.ValidateJWT(accessToken, s.tokenSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get token from header: %v - RefreshToken", err)
	}

	newAccessToken, err := auth.MakeJWT(userID, s.tokenSecret, time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create new access token: %v - RefreshToken", err)
	}

	refreshTokenParams := database.RefreshTokenParams{
		Token:      newAccessToken,
		UserID:     userID,
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
