package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

// Config holds the configuration for Redis connection
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisConfig creates a new config for Redis
func NewRedisConfig(password string) *Config {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // fallback for local development
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	return &Config{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       0,
	}
}

var (
	// Client is a global Redis client instance
	Client *redis.Client
	// Ctx is the background context for Redis operations
	Ctx = context.Background()
)

// InitRedisClient initializes the Redis client with the provided configuration
func InitRedisClient(cfg *Config) {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := Client.Ping().Result()
	if err != nil {
		panic("Не удалось подключиться к Redis: " + err.Error())
	}
}
