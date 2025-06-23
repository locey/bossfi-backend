package redis

import (
	"context"
	"fmt"
	"time"

	"bossfi-blockchain-backend/pkg/config"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client

// Connect 连接Redis
func Connect(cfg *config.Config) (*redis.Client, error) {
	err := InitRedis(&cfg.Redis)
	if err != nil {
		return nil, err
	}
	return Client, nil
}

func InitRedis(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolTimeout:  30 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}

	return nil
}

func GetClient() *redis.Client {
	return Client
}

func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
