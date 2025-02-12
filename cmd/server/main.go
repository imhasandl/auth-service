package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq" // Import the postgres driver

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"github.com/imhasandl/grpc-go/internal/auth"
	"github.com/imhasandl/grpc-go/internal/database"
	pb "github.com/imhasandl/grpc-go/internal/protos"
	"github.com/joho/godotenv"
)

type server struct {
		pb.UnimplementedAuthServiceServer
		db *database.Queries
		tokenSecret string
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
		ID: uuid.New(),
		Email: req.GetEmail(),
		Password: string(hashedPassword),
		Username:  req.GetUsername(),
	}

	user, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create user: %v - Register", err)
	}

	createdAtProto := timestamppb.New(user.CreatedAt) // Converts time.Time type into timespamppb
	updatedAtProto := timestamppb.New(user.UpdatedAt) // Converts time.Time type into timespamppb

	return &pb.RegisterResponse{
		User: &pb.User{
			Id: user.ID.String(),
			CreatedAt: createdAtProto,
			UpdatedAt: updatedAtProto,
			Email: user.Email,
			Username: user.Username,
			IsPremium: user.IsPremium,
		},
	}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userParams := database.GetUserByIdentifierParams{
		Email: req.GetIdentifier(),
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
			Id: user.ID.String(),
			CreatedAt: createdAtProto,
			UpdatedAt: updatedAtProto,
			Email: user.Email,
			Username: user.Username,
			IsPremium: user.IsPremium,
		},
		Token: accessToken,
	}, nil
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("Set Port in env")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("Set db connection in env")
	}

	tokenSecret := os.Getenv("TOKEN_SECRET")
	if tokenSecret == "" {
		log.Fatalf("Set db connection in env")
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)
	defer dbConn.Close()

	server := &server{
		pb.UnimplementedAuthServiceServer{},
		dbQueries,
		tokenSecret,
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, server)

	reflection.Register(s)
	log.Printf("Server listening on %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to lister: %v", err)
	}
}
