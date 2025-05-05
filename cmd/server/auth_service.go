package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	auth "github.com/imhasandl/auth-service/cmd/auth"
	"github.com/imhasandl/auth-service/cmd/helper"
	"github.com/imhasandl/auth-service/internal/database"
	pb "github.com/imhasandl/auth-service/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DBQuerier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUserByIdentifier(ctx context.Context, arg GetUserByIdentifierParams) (User, error)
	VerifyUser(ctx context.Context, email string) error
	StoreVerificationCode(ctx context.Context, arg StoreVerificationCodeParams) error
	SendVerifyCodeAgain(ctx context.Context, arg SendVerifyCodeAgainParams) error
	RefreshToken(ctx context.Context, arg RefreshTokenParams) (RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (RefreshToken, error)
	DeleteTokenByToken(ctx context.Context, token string) error
	DeleteTokenByUserID(ctx context.Context, userID uuid.UUID) error
	// Add other methods as needed
}

// Server implements the AuthService gRPC interface
type Server struct {
	pb.UnimplementedAuthServiceServer
	db          *database.Queries
	tokenSecret string
	email       string
	emailSecret string
}

// NewServer creates and initializes a new AuthService server instance
func NewServer(db DBQuerier, tokenSecret, email, emailSecret string) *Server {
	return &Server{
		pb.UnimplementedAuthServiceServer{},
		db,
		tokenSecret,
		email,
		emailSecret,
	}
}

// Register handles user registration by validating input data, creating a new user record,
// generating a verification code, and sending it to the user's email for account verification.
// It returns the created user information on success or an appropriate error on failure.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
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

// VerifyEmail validates the verification code provided by the user against the one stored in the database.
// If the code is valid and not expired, it marks the user's email as verified.
// It returns a success response or an appropriate error on failure.
func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	userParams := database.GetUserByIdentifierParams{
		Email:    req.GetEmail(),
		Username: "",
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

	if user.VerificationExpireTime.Before(time.Now()) {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.DeadlineExceeded, "verification code expired - VerifyEmail", nil)
	}

	return &pb.VerifyEmailResponse{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}

// SendVerifyCodeAgain generates a new verification code for a user and sends it to their email.
// This is used when the original verification code expired or was lost.
// It returns a success response or an appropriate error on failure.
func (s *Server) SendVerifyCodeAgain(ctx context.Context, req *pb.SendVerifyCodeAgainRequest) (*pb.SendVerifyCodeAgainResponse, error) {
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

// Login authenticates a user using their email/username and password.
// On successful authentication, it generates JWT access and refresh tokens.
// It returns the user information along with the tokens or an appropriate error on failure.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
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

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't create refresh token - Login", err)
	}

	refreshTokenParams := database.RefreshTokenParams{
		Token:      refreshToken,
		UserID:     user.ID,
		ExpiryTime: time.Now().Add(7 * 24 * time.Hour),
	}

	_, err = s.db.RefreshToken(ctx, refreshTokenParams)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't store refresh token - Login", err)
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
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken validates a refresh token and issues a new access token and refresh token pair.
// It invalidates the old refresh token to implement token rotation security.
// It returns the new tokens or an appropriate error if the token is invalid or expired.
func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	refreshToken := req.GetRefreshToken()
	if refreshToken == "" {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.InvalidArgument, "refresh token is required - RefreshToken", nil)
	}

	storedToken, err := s.db.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't get refresh token - RefreshToken", err)
	}

	// Verify token is not expired
	if time.Now().After(storedToken.ExpiryTime) {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Unauthenticated, "refresh token expired - RefreshToken", nil)
	}

	newAccessToken, err := auth.MakeJWT(storedToken.UserID, s.tokenSecret, time.Hour)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't create new access token - RefreshToken", err)
	}

	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't create new refresh token - RefreshToken", err)
	}

	err = s.db.DeleteTokenByUserID(ctx, storedToken.UserID)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't delete old token - RefreshToken", err)
	}

	refreshTokenParams := database.RefreshTokenParams{
		Token:      newRefreshToken,
		UserID:     storedToken.UserID,
		ExpiryTime: time.Now().Add(7 * 24 * time.Hour),
	}

	_, err = s.db.RefreshToken(ctx, refreshTokenParams)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't store refresh token - RefreshToken", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiryTime:   timestamppb.New(time.Now().Add(time.Hour)),
	}, nil
}

// Logout invalidates a user's refresh token, effectively ending their session.
// It deletes the token from the database to prevent its future use.
// It returns a success response or an appropriate error on failure.
func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	refreshToken := req.GetRefreshToken()
	if refreshToken == "" {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.InvalidArgument, "refresh token is required - Logout", nil)
	}

	err := s.db.DeleteTokenByToken(ctx, refreshToken)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.Internal, "can't delete token - Logout", err)
	}

	return &pb.LogoutResponse{
		Success: true,
		Message: "User logged out complete",
	}, nil
}
