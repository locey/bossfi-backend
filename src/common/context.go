package common

import (
	"bossfi-backend/src/core/config"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Ctx = Context{}

type Context struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Pool
	Log    *zap.Logger
}
