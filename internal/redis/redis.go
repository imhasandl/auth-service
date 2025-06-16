package redis

import (
	"fmt"
	"time"
)

// Token type constants used for storing different types of authentication tokens
const (
	AccessToken  = "access"
	RefreshToken = "refresh"
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
func GetVerificationCode(email string) (int, error) {
	key := fmt.Sprintf("verification:%s", email)
	result, err := Client.Get(key).Int()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// DeleteVerificationCode удаляет верификационный код
func DeleteVerificationCode(email string) error {
	key := fmt.Sprintf("verification:%s", email)
	return Client.Del(key).Err()
}

// SaveAccessToken stores token
func SaveAccessToken(userID, token string, expiration time.Duration) error {
	key := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	return Client.Set(key, token, expiration).Err()
}

// GetAccessToken gets token using it's key
func GetAccessToken(userID string) (string, error) {
	key := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	return Client.Get(key).Result()
}

// DeleteAccessToken removes a user's access token from Redis
func DeleteAccessToken(userID string) error {
	key := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	return Client.Del(key).Err()
}

// SaveRefreshToken stores a user's refresh token in Redis with the specified expiration time
func SaveRefreshToken(userID, token string, expiration time.Duration) error {
	key := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	return Client.Set(key, token, expiration).Err()
}

// GetRefreshToken retrieves a user's refresh token from Redis
func GetRefreshToken(userID string) (string, error) {
	key := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	return Client.Get(key).Result()
}

// DeleteRefreshToken removes a user's refresh token from Redis
func DeleteRefreshToken(userID string) error {
	key := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	return Client.Del(key).Err()
}

// DeleteAllUserTokens removes all tokens (access and refresh) associated with a user from Redis
func DeleteAllUserTokens(userID string) error {
	accessKey := fmt.Sprintf("user:%s:%s_token", userID, AccessToken)
	refreshKey := fmt.Sprintf("user:%s:%s_token", userID, RefreshToken)

	return Client.Del(accessKey, refreshKey).Err()
}
