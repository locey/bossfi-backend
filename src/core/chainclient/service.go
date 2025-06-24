package chainclient

import (
	"bossfi-backend/src/common/chain"
	"bossfi-backend/src/core/chainclient/evm"
	"errors"
)

type ChainClient interface {
	Client() interface{}
}

func New(chainId int, endpoint string) (ChainClient, error) {
	switch chainId {
	case chain.EthChainID, chain.OptimismChainID, chain.SepoliaChainID:
		return evm.New(endpoint)
	default:
		return nil, errors.New("unsupported chain id")
	}
}
