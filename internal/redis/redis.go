package redis

import (
	"time"
)

// SaveToken stores a user token in Redis with the specified expiration time
func SaveToken(userID string, token string, expirationTime time.Duration) error {
	return Client.Set("user:token:"+userID, token, expirationTime).Err()
}

// GetToken retrieves a user's token from Redis
func GetToken(userID string) (string, error) {
	return Client.Get("user:token:" + userID).Result()
}

// DeleteToken removes a user's token from Redis
func DeleteToken(userID string) error {
	return Client.Del("user:token:" + userID).Err()
}
