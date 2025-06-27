package services

import (
	"context"
	"math/big"
	"time"

	"bossfi-backend/api/models"
	"bossfi-backend/config"
	"bossfi-backend/db/database"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type BlockchainService struct {
	client       *ethclient.Client
	contractAddr common.Address
}

func NewBlockchainService() *BlockchainService {
	cfg := config.AppConfig.Blockchain

	var client *ethclient.Client
	var err error

	if cfg.RPCURL != "" {
		client, err = ethclient.Dial(cfg.RPCURL)
		if err != nil {
			logrus.Errorf("Failed to connect to blockchain: %v", err)
		} else {
			logrus.Info("Blockchain client connected successfully")
		}
	}

	var contractAddr common.Address
	if cfg.ContractAddress != "" {
		contractAddr = common.HexToAddress(cfg.ContractAddress)
	}

	return &BlockchainService{
		client:       client,
		contractAddr: contractAddr,
	}
}

// GetBlockNumber 获取最新区块号
func (bs *BlockchainService) GetBlockNumber() (*big.Int, error) {
	if bs.client == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blockNum, err := bs.client.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	return big.NewInt(int64(blockNum)), nil
}

// GetBalance 获取钱包余额
func (bs *BlockchainService) GetBalance(address string) (*big.Int, error) {
	if bs.client == nil {
		return big.NewInt(0), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	account := common.HexToAddress(address)
	return bs.client.BalanceAt(ctx, account, nil)
}

// SyncUserBalances 同步用户余额
func (bs *BlockchainService) SyncUserBalances() error {
	if bs.client == nil {
		logrus.Warn("Blockchain client not initialized, skipping balance sync")
		return nil
	}

	logrus.Info("Starting user balance synchronization...")

	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		logrus.Errorf("Failed to fetch users: %v", err)
		return err
	}

	for _, user := range users {
		balance, err := bs.GetBalance(user.WalletAddress)
		if err != nil {
			logrus.Errorf("Failed to get balance for user %s: %v", user.WalletAddress, err)
			continue
		}

		// 将 wei 转换为 ether (假设这里存储的是 ETH 余额)
		// 这里需要根据实际的 token 合约进行调整
		balanceFloat := new(big.Float).SetInt(balance)
		ethBalance, _ := balanceFloat.Quo(balanceFloat, big.NewFloat(1e18)).Float64()

		// 更新用户余额
		updateData := map[string]interface{}{
			"boss_balance": ethBalance,
			"updated_at":   time.Now(),
		}

		if err := database.DB.Model(&user).Updates(updateData).Error; err != nil {
			logrus.Errorf("Failed to update balance for user %s: %v", user.WalletAddress, err)
		} else {
			logrus.Debugf("Updated balance for user %s: %f", user.WalletAddress, ethBalance)
		}
	}

	logrus.Info("User balance synchronization completed")
	return nil
}

// SyncBlockchainData 同步区块链数据（主要入口）
func (bs *BlockchainService) SyncBlockchainData() error {
	logrus.Info("Starting blockchain data synchronization...")

	// 获取最新区块号
	blockNumber, err := bs.GetBlockNumber()
	if err != nil {
		logrus.Errorf("Failed to get latest block number: %v", err)
	} else if blockNumber != nil {
		logrus.Infof("Latest block number: %s", blockNumber.String())
	}

	// 同步用户余额
	if err := bs.SyncUserBalances(); err != nil {
		logrus.Errorf("Failed to sync user balances: %v", err)
		return err
	}

	// 这里可以添加更多的同步逻辑，比如：
	// - 同步质押信息
	// - 同步奖励信息
	// - 同步交易历史
	// - 同步合约事件

	logrus.Info("Blockchain data synchronization completed")
	return nil
}

// GetUserStakingInfo 获取用户质押信息（示例）
func (bs *BlockchainService) GetUserStakingInfo(walletAddress string) (map[string]interface{}, error) {
	// 这里应该调用智能合约获取真实的质押信息
	// 目前返回模拟数据

	if bs.client == nil {
		return map[string]interface{}{
			"staked_amount":  0,
			"reward_balance": 0,
		}, nil
	}

	// 实际实现中，这里应该：
	// 1. 调用智能合约的 stakingInfo(address) 方法
	// 2. 解析返回的数据
	// 3. 返回结构化的质押信息

	return map[string]interface{}{
		"staked_amount":  0, // 从合约获取
		"reward_balance": 0, // 从合约获取
	}, nil
}

// Close 关闭区块链客户端连接
func (bs *BlockchainService) Close() {
	if bs.client != nil {
		bs.client.Close()
		logrus.Info("Blockchain client connection closed")
	}
}
