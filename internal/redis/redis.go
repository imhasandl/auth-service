package redis

import (
	"time"
)

func SaveToken(userID string, token string, expirationTime time.Duration) error {
	return Client.Set("user:token:" + userID, token, expirationTime).Err()
}

func GetToken(userID string) (string, error) {
	return Client.Get("user:token:" + userID).Result()
}

func DeleteToken(userID string) error {
	return Client.Del("user:token:" + userID).Err()
}
