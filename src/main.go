package main

import "bossfi-backend/src/core"

const (
	// ConfigFile 配置文件路径
	ConfigFile = "config.toml"
)

func main() {
	core.Start(ConfigFile)
}
