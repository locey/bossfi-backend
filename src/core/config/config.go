package config

import (
	"flag"
	"github.com/spf13/viper"
	"strings"
)

var Conf *Config

type Config struct {
	App   AppConfig
	Pgsql PgsqlConfig
	Redis RedisConfig
}

type AppConfig struct {
	Name    string `toml:"name" json:"name"`
	Port    string `toml:"port" json:"port"`
	Version string `toml:"version" json:"version"`
}

type PgsqlConfig struct {
	Host     string `toml:"host" json:"host"`
	Port     string `toml:"port" json:"port"`
	Username string `toml:"username" json:"username"`
	Password string `toml:"password" json:"password"`
	Database string `toml:"database" json:"database"`
}

type RedisConfig struct {
	Host        string `toml:"host" json:"host"`
	Port        string `toml:"port" json:"port"`
	Password    string `toml:"password" json:"password"`
	Db          int    `toml:"db" json:"db"`
	MaxIdle     int    `toml:"max_idle" json:"maxIdle"`
	MaxActive   int    `toml:"max_active" json:"maxActive"`
	IdleTimeout int    `toml:"idle_timeout" json:"idleTimeout"`
}

// InitConfig 初始化配置
func InitConfig(configPath string) *Config {
	conf := flag.String("conf", configPath, "conf file path")
	flag.Parse()
	c, err := UnmarshalConfig(*conf)
	if err != nil {
		panic(err)
	}
	Conf = c
	return c
}

// UnmarshalConfig 解析配置文件
func UnmarshalConfig(configFilePath string) (*Config, error) {
	viper.SetConfigFile(configFilePath)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("BOSSFI")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	config, err := DefaultConfig()
	if err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func DefaultConfig() (*Config, error) {
	return &Config{}, nil
}
