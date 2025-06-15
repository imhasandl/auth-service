package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedisConfig(password string) *RedisConfig {
	return &RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: password,
		DB:       0,
	}
}

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func InitRedisClient(cfg *RedisConfig) {
	Client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := Client.Ping().Result()
	if err != nil {
		panic("Не удалось подключиться к Redis: " + err.Error())
	}
}
