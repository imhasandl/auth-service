package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	auth "github.com/imhasandl/auth-service/cmd/auth"
	"github.com/imhasandl/auth-service/cmd/helper"
	"github.com/imhasandl/auth-service/internal/database"
	pb "github.com/imhasandl/auth-service/protos"
	postService "github.com/imhasandl/post-service/cmd/helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	db          *database.Queries
	tokenSecret string
	email       string
	emailSecret string
}

func NewServer(db *database.Queries, tokenSecret, email, emailSecret string) *server {
	return &server{
		pb.UnimplementedAuthServiceServer{},
		db,
		tokenSecret,
		email,
		emailSecret,
	}
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(req.GetUsername()) < 5 {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "username should be 5 characters long", nil)
	}

	hashedPassword, err := auth.HashPassword(req.GetPassword())
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to hash password - Register", err)
	}

	verificationCode, err := auth.GenerateVerificationCode()
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to generate verification code - Register", err)
	}

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
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't create user - Register", err)
	}

	verifyParams := database.StoreVerificationCodeParams{
		VerificationCode: verificationCode,
		ID:               user.ID,
	}

	err = s.db.StoreVerificationCode(ctx, verifyParams)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to store verification code - Register", err)
	}

	err = auth.SendVerificationEmail(req.GetEmail(), s.email, s.emailSecret, verificationCode)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to send verification email - Register", err)
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
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't get user with identifier - VerifyEmail", err)
	}

	if user.IsVerified {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.AlreadyExists, "email already verified - VerifyEmail", nil)
	}

	if user.VerificationCode != req.GetVerificationCode() {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Unauthenticated, "invalid verification code - VerifyEmail", nil)
	}

	err = s.db.VerifyUser(ctx, req.GetEmail())
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to verify user - VerifyEmail", err)
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
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't get user with identifier - SendVerifyCodeAgain", err)
	}

	newVerifyCode, err := auth.GenerateVerificationCode()
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to generate verification code - SendVerifyCodeAgain", err)
	}

	sendVerifyAgainParams := database.SendVerifyCodeAgainParams{
		VerificationCode: newVerifyCode,
		ID:               user.ID,
	}

	err = s.db.SendVerifyCodeAgain(ctx, sendVerifyAgainParams)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to send verification code again - SendVerifyCodeAgain", err)
	}

	err = auth.SendVerificationEmail(req.GetEmail(), s.email, s.emailSecret, newVerifyCode)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "failed to send verification email - SendVerifyCodeAgain", err)
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
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't get user with identifier - Login", err)
	}

	err = auth.CheckPassword(user.Password, req.GetPassword())
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Unauthenticated, "invalid credentials - Login", err)
	}

	accessToken, err := auth.MakeJWT(user.ID, s.tokenSecret, time.Hour)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't create token - Login", err)
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
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Unauthenticated, "invalid token - RefreshToken", err)
	}

	userID, err := postService.ValidateJWT(accessToken, s.tokenSecret)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't get token from header - RefreshToken", err)
	}

	newAccessToken, err := auth.MakeJWT(userID, s.tokenSecret, time.Hour)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't create new access token - RefreshToken", err)
	}

	refreshTokenParams := database.RefreshTokenParams{
		Token:      newAccessToken,
		UserID:     userID,
		ExpiryTime: time.Now().Add(24 * time.Hour),
	}

	refreshToken, err := s.db.RefreshToken(ctx, refreshTokenParams)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't store refresh token - RefreshToken", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newAccessToken,
		ExpiryTime:   timestamppb.New(refreshToken.ExpiryTime),
	}, nil
}

func (s *server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	accessToken, err := postService.GetBearerTokenFromGrpc(ctx)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Unauthenticated, "invalid token - Logout", err)
	}

	userID, err := postService.ValidateJWT(accessToken, s.tokenSecret)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't get token from header - Logout", err)
	}

	err = s.db.DeleteToken(ctx, userID)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't delete token - Logout", err)
	}

	return &pb.LogoutResponse{
		Success: true,
		Message: "User logged out complete",
	}, nil
}
