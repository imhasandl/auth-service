package helper

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// EnvConfig contains all the environment variables required for the application
type EnvConfig struct {
	Port        string
	DBURL       string
	Email       string
	EmailSecret string
	TokenSecret string
	RedisSecret string
}

// GetENVSecrets loads environment variables from .env file and returns the configuration
func GetENVSecrets() EnvConfig {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Error loading .env file")
	}

	config := EnvConfig{
		Port:        os.Getenv("PORT"),
		DBURL:       os.Getenv("DB_URL"),
		Email:       os.Getenv("EMAIL"),
		EmailSecret: os.Getenv("EMAIL_SECRET"),
		TokenSecret: os.Getenv("TOKEN_SECRET"),
		RedisSecret: os.Getenv("REDIS_SECRET"),
	}

	if config.Port == "" {
		log.Fatalf("Set Port in env")
	}
	if config.DBURL == "" {
		log.Fatalf("Set db connection in env")
	}
	if config.Email == "" {
		log.Fatal("Set up Email in env")
	}
	if config.EmailSecret == "" {
		log.Fatalf("Set up Email Secret in env")
	}
	if config.TokenSecret == "" {
		log.Fatalf("Set token secret in env")
	}
	if config.RedisSecret == "" {
		log.Fatalf("Set redis password in .env file")
	}

	return config
}
