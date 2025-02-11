package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/imhasandl/grpc-go/internal/protos"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file:", err)  // Include err for better debugging
  }  

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("Set Port in env")
	}

	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewAuthServiceClient(conn)

	// Register method
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	registerReq := &pb.RegisterRequest{Email: "tet@exame.com", Password: "password123", Username: "tetuer"}
	registerResp, err := c.Register(ctx, registerReq)
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}
	log.Printf("Register response: %v", registerResp)

	// Login method
	loginReq := &pb.LoginRequest{Identifier: "test@example.com", Password: "password123"}
	loginResp, err := c.Login(ctx, loginReq)
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Login response: %v", loginResp)
}