package redis

import (
	"context"
	"fmt"
	"time"

	"bossfi-backend/config"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis() error {
	cfg := config.AppConfig.Redis

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		logrus.Errorf("Failed to connect to Redis: %v", err)
		return err
	}

	logrus.Info("Redis connected successfully")
	return nil
}

func GetRedisClient() *redis.Client {
	return RedisClient
}

// SetNonce 设置用户的 nonce 值
func SetNonce(walletAddress, nonce string) error {
	key := fmt.Sprintf("nonce:%s", walletAddress)
	return RedisClient.Set(ctx, key, nonce, 5*time.Minute).Err()
}

// GetNonce 获取用户的 nonce 值
func GetNonce(walletAddress string) (string, error) {
	key := fmt.Sprintf("nonce:%s", walletAddress)
	return RedisClient.Get(ctx, key).Result()
}

// DeleteNonce 删除用户的 nonce 值
func DeleteNonce(walletAddress string) error {
	key := fmt.Sprintf("nonce:%s", walletAddress)
	return RedisClient.Del(ctx, key).Err()
}

// SetUserSession 设置用户会话
func SetUserSession(userID, token string) error {
	key := fmt.Sprintf("session:%s", userID)
	return RedisClient.Set(ctx, key, token, config.AppConfig.JWT.ExpireHours).Err()
}

// GetUserSession 获取用户会话
func GetUserSession(userID string) (string, error) {
	key := fmt.Sprintf("session:%s", userID)
	return RedisClient.Get(ctx, key).Result()
}

// DeleteUserSession 删除用户会话
func DeleteUserSession(userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return RedisClient.Del(ctx, key).Err()
}
