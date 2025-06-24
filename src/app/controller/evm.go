package controller

import (
	"bossfi-backend/src/common/chain"
	"bossfi-backend/src/core/chainclient/domain"
	"bossfi-backend/src/core/ctx"
	"bossfi-backend/src/core/result"
	"context"
	"github.com/gin-gonic/gin"
	"math/big"
	"strconv"
)

// GetBlockByNum curl "http://localhost:8000/api/v1/evm/get_block_by_num/8615565"
func GetBlockByNum(c *gin.Context) {
	blockNumString := c.Params.ByName("block_num")
	if blockNumString == "" {
		result.Error(c, result.InvalidParameter)
		return
	}
	blockNum, err := strconv.ParseInt(blockNumString, 10, 64)
	if err != nil {
		result.Error(c, result.InvalidParameter)
		return
	}

	blockNewHeader := big.NewInt(blockNum)

	client := ctx.GetEvmClient(chain.SepoliaChainID)
	block, err := client.BlockByNumber(context.Background(), blockNewHeader)
	if err != nil {
		result.Error(c, result.EthereumError)
		return
	}

	result.OK(c, domain.ToBlock(block))
}
