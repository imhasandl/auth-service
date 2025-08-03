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
	RedisHost   string
	RedisPort   string
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
		RedisHost:   os.Getenv("REDIS_HOST"),
		RedisPort:   os.Getenv("REDIS_PORT"),
		RedisSecret: os.Getenv("REDIS_SECRET"),
	}

	if config.Port == "" {
		log.Fatal("Set Port in env")
	}
	if config.DBURL == "" {
		log.Fatal("Set db connection in env")
	}
	if config.Email == "" {
		log.Fatal("Set up Email in env")
	}
	if config.EmailSecret == "" {
		log.Fatal("Set up Email Secret in env")
	}
	if config.TokenSecret == "" {
		log.Fatal("Set token secret in env")
	}
	if config.RedisHost == "" {
		log.Fatal("Set redis how in .env file")
	}
	if config.RedisPort == "" {
		log.Fatal("Set redis port in .env file")
	}
	if config.RedisSecret == "" {
		log.Fatal("Set redis password in .env file")
	}
	return config
}
