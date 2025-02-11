package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"golang.org/x/crypto/bcrypt" 
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"github.com/imhasandl/grpc-go/internal/database"
	pb "github.com/imhasandl/grpc-go/internal/protos"
	"github.com/joho/godotenv"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	db *database.Queries
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(req.Username) < 5 {
		return nil, status.Errorf(codes.Internal, "username should be 5 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v - Register", err)
	}
	
	userParams := database.CreateUserParams{
		ID: uuid.New(),
		Email: req.Email,
		Password: string(hashedPassword),
		Username:  req.Username,
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

// func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

// }

func main() {
	if err := godotenv.Load(); err != nil {
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
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, server)

	reflection.Register(s)
	log.Printf("Server listening on %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to lister: %v", err)
	}
}
