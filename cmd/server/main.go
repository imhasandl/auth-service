package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/imhasandl/grpc-go/internal/database"
	pb "github.com/imhasandl/grpc-go/internal/protos"
	"github.com/joho/godotenv"
)

type server struct {
	pb.UnimplementedAuthServiceServer	
	db *database.Queries
}


// func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {

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

	server := server{
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
