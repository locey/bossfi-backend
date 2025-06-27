package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Server     ServerConfig
	Blockchain BlockchainConfig
	Cron       CronConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpireHours time.Duration
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type BlockchainConfig struct {
	RPCURL          string
	ContractAddress string
	PrivateKey      string
}

type CronConfig struct {
	Enabled                bool
	BlockchainSyncInterval string
}

var AppConfig *Config

func Init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	cronEnabled, _ := strconv.ParseBool(getEnv("CRON_ENABLED", "true"))

	AppConfig = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "bossfi"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-super-secret-key"),
			ExpireHours: time.Duration(expireHours) * time.Hour,
		},
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Blockchain: BlockchainConfig{
			RPCURL:          getEnv("BLOCKCHAIN_RPC_URL", ""),
			ContractAddress: getEnv("CONTRACT_ADDRESS", ""),
			PrivateKey:      getEnv("PRIVATE_KEY", ""),
		},
		Cron: CronConfig{
			Enabled:                cronEnabled,
			BlockchainSyncInterval: getEnv("BLOCKCHAIN_SYNC_INTERVAL", "*/5 * * * *"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
