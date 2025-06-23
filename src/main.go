package main

import (
	"bossfi-backend/src/core"
)

const (
	// ConfigPath 配置文件路径
	ConfigPath = "./config/config.toml"
)

func main() {
	core.Start(ConfigPath)
}
