package ctx

import (
	"bossfi-backend/src/core/chainclient"
	"bossfi-backend/src/core/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Ctx = Context{}

type Context struct {
	Config   *config.Config
	DB       *gorm.DB
	Redis    *redis.Pool
	Log      *zap.Logger
	ChainMap map[int]*chainclient.ChainClient
	Gin      *gin.Engine
}

func GetEvmClient(chainId int) *ethclient.Client {
	client := *Ctx.ChainMap[chainId]
	return client.Client().(*ethclient.Client)
}
