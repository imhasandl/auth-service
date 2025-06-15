package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq" // Import the postgres driver

	"github.com/imhasandl/auth-service/cmd/helper"
	server "github.com/imhasandl/auth-service/cmd/server"
	"github.com/imhasandl/auth-service/internal/database"
	"github.com/imhasandl/auth-service/internal/redis"
	pb "github.com/imhasandl/auth-service/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	envConfig := helper.GetENVSecrets()

	lis, err := net.Listen("tcp", envConfig.Port)
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	dbConn, err := sql.Open("postgres", envConfig.DBURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)
	defer dbConn.Close()

	redisConfig := redis.NewRedisConfig(envConfig.RedisSecret)
	redis.InitRedisClient(redisConfig)

	server := server.NewServer(dbQueries, envConfig.TokenSecret, envConfig.Email, envConfig.EmailSecret)

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, server)

	reflection.Register(s)
	log.Printf("Server listening on %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to lister: %v", err)
	}
}
