package redis

import (
	"fmt"
	"time"
)

// SaveToken stores a user token in Redis with the specified expiration time
func SaveToken(userID, token string, expirationTime time.Duration) error {
	return Client.Set(userID, token, expirationTime).Err()
}

// GetToken retrieves a user's token from Redis
func GetToken(userID string) (string, error) {
	return Client.Get(userID).Result()
}

// DeleteToken removes a user's token from Redis
func DeleteToken(userID string) error {
	return Client.Del(userID).Err()
}

// CacheVerificationCode сохраняет верификационный код для email
func CacheVerificationCode(email string, code int32, expiration time.Duration) error {
	key := fmt.Sprintf("verification:%s", email)
	return Client.Set(key, code, expiration).Err()
}

// GetVerificationCode получает верификационный код для email
func GetVerificationCode(email string) (int32, error) {
	key := fmt.Sprintf("verification:%s", email)
	result, err := Client.Get(key).Int()
	if err != nil {
		return 0, err
	}
	return int32(result), nil
}

// DeleteVerificationCode удаляет верификационный код
func DeleteVerificationCode(email string) error {
	key := fmt.Sprintf("verification:%s", email)
	return Client.Del(key).Err()
}
