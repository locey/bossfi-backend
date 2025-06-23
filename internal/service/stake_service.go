package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/domain/stake"
	"bossfi-blockchain-backend/internal/repository"
	"bossfi-blockchain-backend/pkg/config"
	"bossfi-blockchain-backend/pkg/logger"
)

// StakeService 质押服务接口
type StakeService interface {
	// 质押管理
	CreateStake(ctx context.Context, userID string, req *CreateStakeRequest) (*stake.Stake, error)
	GetStake(ctx context.Context, stakeID string) (*stake.Stake, error)
	GetUserStakes(ctx context.Context, userID string, page, pageSize int) (*ListStakesResponse, error)
	RequestUnstake(ctx context.Context, userID string, stakeID string) (*stake.Stake, error)
	CompleteUnstake(ctx context.Context, userID string, stakeID string) (*stake.Stake, error)

	// 奖励管理
	DistributeRewards(ctx context.Context) (*DistributeRewardsResponse, error)
	ClaimRewards(ctx context.Context, userID string) (*ClaimRewardsResponse, error)

	// 统计信息
	GetStakeStats(ctx context.Context) (*repository.StakeStats, error)
	GetUserStakeStats(ctx context.Context, userID string) (*repository.UserStakeStats, error)
}

// 请求和响应结构
type CreateStakeRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type ListStakesResponse struct {
	Stakes     []*stake.Stake `json:"stakes"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type DistributeRewardsResponse struct {
	ProcessedCount int             `json:"processed_count"`
	TotalReward    decimal.Decimal `json:"total_reward"`
	FailedCount    int             `json:"failed_count"`
}

type ClaimRewardsResponse struct {
	ClaimedAmount decimal.Decimal `json:"claimed_amount"`
	NewBalance    decimal.Decimal `json:"new_balance"`
}

// stakeService 质押服务实现
type stakeService struct {
	stakeRepo repository.StakeRepository
	userRepo  repository.UserRepository
	cfg       *config.Config
	logger    *logger.Logger
}

// NewStakeService 创建质押服务
func NewStakeService(
	stakeRepo repository.StakeRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	logger *logger.Logger,
) StakeService {
	return &stakeService{
		stakeRepo: stakeRepo,
		userRepo:  userRepo,
		cfg:       cfg,
		logger:    logger,
	}
}

// CreateStake 创建质押
func (s *stakeService) CreateStake(ctx context.Context, userID string, req *CreateStakeRequest) (*stake.Stake, error) {
	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取最小质押金额配置
	minStakeAmount, err := s.getMinStakeAmount()
	if err != nil {
		return nil, fmt.Errorf("failed to get min stake amount: %w", err)
	}

	// 验证质押金额
	if req.Amount.LessThan(minStakeAmount) {
		return nil, stake.ErrMinStakeAmountNotMet
	}

	// 检查用户余额是否充足
	if !u.HasSufficientBalance(req.Amount) {
		return nil, stake.ErrInsufficientBalance
	}

	// 创建质押记录
	stakeRecord := &stake.Stake{
		UserID: userID,
		Amount: req.Amount,
		Status: stake.StakeStatusActive,
	}

	// 保存质押记录
	if err := s.stakeRepo.Create(ctx, stakeRecord); err != nil {
		return nil, fmt.Errorf("failed to create stake: %w", err)
	}

	// 更新用户余额
	if err := u.SubBossBalance(req.Amount); err != nil {
		s.logger.Error("Failed to deduct balance", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to deduct balance: %w", err)
	}

	u.AddStakedAmount(req.Amount)

	if err := s.userRepo.Update(ctx, u); err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	s.logger.Info("Stake created successfully", zap.String("user_id", userID), zap.String("stake_id", stakeRecord.ID), zap.String("amount", req.Amount.String()))

	return stakeRecord, nil
}

// GetStake 获取质押详情
func (s *stakeService) GetStake(ctx context.Context, stakeID string) (*stake.Stake, error) {
	return s.stakeRepo.GetByID(ctx, stakeID)
}

// GetUserStakes 获取用户质押列表
func (s *stakeService) GetUserStakes(ctx context.Context, userID string, page, pageSize int) (*ListStakesResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	stakes, total, err := s.stakeRepo.GetByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &ListStakesResponse{
		Stakes:     stakes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// RequestUnstake 请求解质押
func (s *stakeService) RequestUnstake(ctx context.Context, userID string, stakeID string) (*stake.Stake, error) {
	// 获取质押记录
	stakeRecord, err := s.stakeRepo.GetByID(ctx, stakeID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if stakeRecord.UserID != userID {
		return nil, stake.ErrStakeNotFound
	}

	// 请求解质押
	if err := stakeRecord.RequestUnstake(); err != nil {
		return nil, err
	}

	// 保存更新
	if err := s.stakeRepo.Update(ctx, stakeRecord); err != nil {
		return nil, fmt.Errorf("failed to update stake: %w", err)
	}

	s.logger.Info("Unstake requested", zap.String("user_id", userID), zap.String("stake_id", stakeID))

	return stakeRecord, nil
}

// CompleteUnstake 完成解质押
func (s *stakeService) CompleteUnstake(ctx context.Context, userID string, stakeID string) (*stake.Stake, error) {
	// 获取质押记录
	stakeRecord, err := s.stakeRepo.GetByID(ctx, stakeID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if stakeRecord.UserID != userID {
		return nil, stake.ErrStakeNotFound
	}

	// 获取解质押延迟天数
	delayDays, err := s.getUnstakeDelayDays()
	if err != nil {
		return nil, fmt.Errorf("failed to get unstake delay: %w", err)
	}

	// 检查是否可以完成解质押
	if !stakeRecord.CanCompleteUnstake(delayDays) {
		return nil, stake.ErrUnstakeDelayNotMet
	}

	// 完成解质押
	if err := stakeRecord.CompleteUnstake(); err != nil {
		return nil, err
	}

	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 返还质押金额和奖励到用户余额
	totalAmount := stakeRecord.Amount.Add(stakeRecord.RewardEarned)
	u.AddBossBalance(totalAmount)

	// 减少质押金额
	if err := u.SubStakedAmount(stakeRecord.Amount); err != nil {
		s.logger.Error("Failed to subtract staked amount", zap.Error(err), zap.String("user_id", userID))
	}

	// 更新用户信息
	if err := s.userRepo.Update(ctx, u); err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	// 保存质押记录更新
	if err := s.stakeRepo.Update(ctx, stakeRecord); err != nil {
		return nil, fmt.Errorf("failed to update stake: %w", err)
	}

	s.logger.Info("Unstake completed", zap.String("user_id", userID), zap.String("stake_id", stakeID), zap.String("amount", totalAmount.String()))

	return stakeRecord, nil
}

// DistributeRewards 分发奖励
func (s *stakeService) DistributeRewards(ctx context.Context) (*DistributeRewardsResponse, error) {
	// 获取可以获得奖励的质押记录
	stakes, err := s.stakeRepo.GetStakesReadyForReward(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stakes ready for reward: %w", err)
	}

	response := &DistributeRewardsResponse{
		ProcessedCount: 0,
		TotalReward:    decimal.Zero,
		FailedCount:    0,
	}

	// 获取年化奖励率
	annualRate, err := s.getAnnualRewardRate()
	if err != nil {
		return nil, fmt.Errorf("failed to get annual reward rate: %w", err)
	}

	// 计算日奖励率
	dailyRate := annualRate.Div(decimal.NewFromInt(365))

	// 为每个质押记录分发奖励
	for _, stakeRecord := range stakes {
		// 计算奖励金额
		rewardAmount := stakeRecord.Amount.Mul(dailyRate)

		// 添加奖励到质押记录
		stakeRecord.AddReward(rewardAmount)

		// 更新质押记录
		if err := s.stakeRepo.Update(ctx, stakeRecord); err != nil {
			s.logger.Error("Failed to update stake with reward", zap.Error(err), zap.String("stake_id", stakeRecord.ID))
			response.FailedCount++
			continue
		}

		// 更新用户奖励余额
		u, err := s.userRepo.GetByID(ctx, stakeRecord.UserID)
		if err != nil {
			s.logger.Error("Failed to get user for reward", zap.Error(err), zap.String("user_id", stakeRecord.UserID))
			response.FailedCount++
			continue
		}

		u.AddRewardBalance(rewardAmount)
		if err := s.userRepo.Update(ctx, u); err != nil {
			s.logger.Error("Failed to update user reward balance", zap.Error(err), zap.String("user_id", stakeRecord.UserID))
			response.FailedCount++
			continue
		}

		response.ProcessedCount++
		response.TotalReward = response.TotalReward.Add(rewardAmount)

		s.logger.Info("Reward distributed", zap.String("user_id", stakeRecord.UserID), zap.String("stake_id", stakeRecord.ID), zap.String("amount", rewardAmount.String()))
	}

	s.logger.Info("Reward distribution completed", zap.Int("processed", response.ProcessedCount), zap.Int("failed", response.FailedCount), zap.String("total_reward", response.TotalReward.String()))

	return response, nil
}

// ClaimRewards 领取奖励
func (s *stakeService) ClaimRewards(ctx context.Context, userID string) (*ClaimRewardsResponse, error) {
	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查是否有奖励可领取
	if u.RewardBalance.IsZero() {
		return &ClaimRewardsResponse{
			ClaimedAmount: decimal.Zero,
			NewBalance:    u.BossBalance,
		}, nil
	}

	// 将奖励余额转移到BOSS余额
	claimedAmount := u.RewardBalance
	u.AddBossBalance(claimedAmount)
	u.RewardBalance = decimal.Zero

	// 更新用户信息
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	s.logger.Info("Rewards claimed", zap.String("user_id", userID), zap.String("amount", claimedAmount.String()))

	return &ClaimRewardsResponse{
		ClaimedAmount: claimedAmount,
		NewBalance:    u.BossBalance,
	}, nil
}

// GetStakeStats 获取质押统计信息
func (s *stakeService) GetStakeStats(ctx context.Context) (*repository.StakeStats, error) {
	return s.stakeRepo.GetStakeStats(ctx)
}

// GetUserStakeStats 获取用户质押统计信息
func (s *stakeService) GetUserStakeStats(ctx context.Context, userID string) (*repository.UserStakeStats, error) {
	return s.stakeRepo.GetUserStakeStats(ctx, userID)
}

// 辅助函数

// getMinStakeAmount 获取最小质押金额
func (s *stakeService) getMinStakeAmount() (decimal.Decimal, error) {
	// 从配置或数据库获取
	minAmount := "0.001" // 默认值
	return decimal.NewFromString(minAmount)
}

// getUnstakeDelayDays 获取解质押延迟天数
func (s *stakeService) getUnstakeDelayDays() (int, error) {
	// 从配置或数据库获取
	delayDays := "7" // 默认值
	return strconv.Atoi(delayDays)
}

// getAnnualRewardRate 获取年化奖励率
func (s *stakeService) getAnnualRewardRate() (decimal.Decimal, error) {
	// 从配置或数据库获取
	rate := "0.10" // 默认10%
	return decimal.NewFromString(rate)
}
