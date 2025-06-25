package main

import (
	"bossfi-backend/src/core"
	_ "bossfi-backend/src/docs"
)

const (
	// ConfigFile 配置文件路径
	ConfigFile = "config.toml"
)

func main() {
	core.Start(ConfigFile)
}
