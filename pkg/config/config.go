package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
	Security   SecurityConfig   `mapstructure:"security"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Debug   bool   `mapstructure:"debug"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type JWTConfig struct {
	AccessSecret       string `mapstructure:"access_secret"`
	RefreshSecret      string `mapstructure:"refresh_secret"`
	AccessTokenExpire  int    `mapstructure:"access_token_expire"`  // 分钟
	RefreshTokenExpire int    `mapstructure:"refresh_token_expire"` // 小时
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type BlockchainConfig struct {
	Network            string `mapstructure:"network"`
	ConfirmationBlocks int    `mapstructure:"confirmation_blocks"`
	GasPrice           int64  `mapstructure:"gas_price"`
	GasLimit           uint64 `mapstructure:"gas_limit"`
}

type SecurityConfig struct {
	RateLimit      int      `mapstructure:"rate_limit"`
	CorsOrigins    []string `mapstructure:"cors_origins"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

var GlobalConfig *Config

func Load(configPath ...string) (*Config, error) {
	// 默认配置文件路径
	path := "configs/config.toml"
	if len(configPath) > 0 && configPath[0] != "" {
		path = configPath[0]
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("toml")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	GlobalConfig = &config
	return &config, nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "bossfi-backend")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.debug", true)

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	// Database defaults
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.database", "bossfi")
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parse_time", true)
	viper.SetDefault("database.loc", "Local")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// JWT defaults
	viper.SetDefault("jwt.access_secret", "your-access-secret-key")
	viper.SetDefault("jwt.refresh_secret", "your-refresh-secret-key")
	viper.SetDefault("jwt.access_token_expire", 60)   // 60 minutes
	viper.SetDefault("jwt.refresh_token_expire", 168) // 7 days

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.filename", "./logs/app.log")
	viper.SetDefault("logger.max_size", 100)
	viper.SetDefault("logger.max_age", 30)
	viper.SetDefault("logger.max_backups", 10)
	viper.SetDefault("logger.compress", true)

	// Security defaults
	viper.SetDefault("security.rate_limit", 100)
	viper.SetDefault("security.cors_origins", []string{"*"})
	viper.SetDefault("security.trusted_proxies", []string{})
}

func Get() *Config {
	return GlobalConfig
}
