package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/logger"
	"bossfi-blockchain-backend/pkg/mreturn"
)

// StakeHandler 质押处理器
type StakeHandler struct {
	stakeService service.StakeService
	logger       *logger.Logger
}

// NewStakeHandler 创建质押处理器
func NewStakeHandler(stakeService service.StakeService, logger *logger.Logger) *StakeHandler {
	return &StakeHandler{
		stakeService: stakeService,
		logger:       logger,
	}
}

// CreateStake 创建质押
// @Summary 创建质押
// @Tags 质押
// @Security BearerAuth
// @Accept json
// @Produce json
// @Router /stakes [post]
func (h *StakeHandler) CreateStake(c *gin.Context) {
	userID := h.getUserID(c)

	var req service.CreateStakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create stake request", zap.Error(err), zap.String("user_id", userID))
		mreturn.BadRequest(c, "Invalid request parameters")
		return
	}

	stake, err := h.stakeService.CreateStake(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create stake", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to create stake")
		return
	}

	c.JSON(http.StatusCreated, mreturn.Response{
		Code:    0,
		Message: "success",
		Data:    stake,
	})
}

// GetStake 获取质押详情
// @Summary 获取质押详情
// @Tags 质押
// @Security BearerAuth
// @Produce json
// @Router /stakes/{id} [get]
func (h *StakeHandler) GetStake(c *gin.Context) {
	stakeID := c.Param("id")
	if stakeID == "" {
		mreturn.BadRequest(c, "stake ID is required")
		return
	}

	stake, err := h.stakeService.GetStake(c.Request.Context(), stakeID)
	if err != nil {
		h.logger.Error("Failed to get stake", zap.Error(err), zap.String("stake_id", stakeID))
		mreturn.NotFound(c, "Stake not found")
		return
	}

	mreturn.Success(c, stake)
}

// GetUserStakes 获取用户质押列表
// @Summary 获取用户质押列表
// @Tags 质押
// @Security BearerAuth
// @Produce json
// @Router /stakes/user/{user_id} [get]
func (h *StakeHandler) GetUserStakes(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		mreturn.BadRequest(c, "user ID is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	user_stakes_response, err := h.stakeService.GetUserStakes(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user stakes", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to get user stakes")
		return
	}

	mreturn.Success(c, user_stakes_response)
}

// RequestUnstake 请求解质押
// @Summary 请求解质押
// @Tags 质押
// @Security BearerAuth
// @Router /stakes/{id}/unstake [post]
func (h *StakeHandler) RequestUnstake(c *gin.Context) {
	userID := h.getUserID(c)
	stakeID := c.Param("id")

	if stakeID == "" {
		mreturn.BadRequest(c, "stake ID is required")
		return
	}

	stake, err := h.stakeService.RequestUnstake(c.Request.Context(), userID, stakeID)
	if err != nil {
		h.logger.Error("Failed to request unstake", zap.Error(err), zap.String("user_id", userID), zap.String("stake_id", stakeID))
		mreturn.BadRequest(c, "Failed to request unstake")
		return
	}

	mreturn.Success(c, stake)
}

// CompleteUnstake 完成解质押
// @Summary 完成解质押
// @Tags 质押
// @Security BearerAuth
// @Router /stakes/{id}/complete [post]
func (h *StakeHandler) CompleteUnstake(c *gin.Context) {
	userID := h.getUserID(c)
	stakeID := c.Param("id")

	if stakeID == "" {
		mreturn.BadRequest(c, "stake ID is required")
		return
	}

	stake, err := h.stakeService.CompleteUnstake(c.Request.Context(), userID, stakeID)
	if err != nil {
		h.logger.Error("Failed to complete unstake", zap.Error(err), zap.String("user_id", userID), zap.String("stake_id", stakeID))
		mreturn.BadRequest(c, "Failed to complete unstake")
		return
	}

	mreturn.Success(c, stake)
}

// ClaimRewards 领取奖励
// @Summary 领取奖励
// @Tags 质押
// @Security BearerAuth
// @Router /stakes/rewards/claim [post]
func (h *StakeHandler) ClaimRewards(c *gin.Context) {
	userID := h.getUserID(c)

	claim_response, err := h.stakeService.ClaimRewards(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to claim rewards", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to claim rewards")
		return
	}

	mreturn.Success(c, claim_response)
}

// GetUserStakeStats 获取用户质押统计信息
// @Summary 获取用户质押统计信息
// @Tags 质押
// @Security BearerAuth
// @Router /stakes/user/{user_id}/stats [get]
func (h *StakeHandler) GetUserStakeStats(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		mreturn.BadRequest(c, "user ID is required")
		return
	}

	stats, err := h.stakeService.GetUserStakeStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user stake stats", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to get user stake stats")
		return
	}

	mreturn.Success(c, stats)
}

// 管理员功能

// DistributeRewards 分发奖励（管理员功能）
// @Summary 分发奖励
// @Tags 管理员
// @Security BearerAuth
// @Router /admin/stakes/distribute [post]
func (h *StakeHandler) DistributeRewards(c *gin.Context) {
	distribute_response, err := h.stakeService.DistributeRewards(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to distribute rewards", zap.Error(err))
		mreturn.InternalServerError(c, "Failed to distribute rewards")
		return
	}

	mreturn.Success(c, distribute_response)
}

// GetStakeStats 获取质押统计信息（管理员功能）
// @Summary 获取质押统计信息
// @Tags 管理员
// @Security BearerAuth
// @Router /admin/stakes/stats [get]
func (h *StakeHandler) GetStakeStats(c *gin.Context) {
	stats, err := h.stakeService.GetStakeStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get stake stats", zap.Error(err))
		mreturn.InternalServerError(c, "Failed to get stake stats")
		return
	}

	mreturn.Success(c, stats)
}

// 辅助函数
func (h *StakeHandler) getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
