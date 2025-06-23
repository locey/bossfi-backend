package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/logger"
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

// 响应函数
func (h *StakeHandler) successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func (h *StakeHandler) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"code":    status,
		"message": message,
	})
}

func (h *StakeHandler) successResponseWithStatus(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// CreateStake 创建质押
// @Summary 创建质押
// @Description 创建新的质押记录
// @Tags 质押
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} object "成功"
// @Failure 400 {object} object "请求参数错误"
// @Failure 401 {object} object "未授权"
// @Failure 500 {object} object "服务器错误"
// @Router /stakes [post]
func (h *StakeHandler) CreateStake(c *gin.Context) {
	userID := h.getUserID(c)

	var req service.CreateStakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create stake request", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	stake, err := h.stakeService.CreateStake(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create stake", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to create stake")
		return
	}

	h.successResponseWithStatus(c, http.StatusCreated, stake)
}

// GetStake 获取质押详情
// @Summary 获取质押详情
// @Description 根据ID获取质押详细信息
// @Tags 质押
// @Security BearerAuth
// @Produce json
// @Param id path string true "质押ID"
// @Success 200 {object} api.Response{data=stake.Stake} "成功"
// @Failure 400 {object} api.Response "请求参数错误"
// @Failure 401 {object} api.Response "未授权"
// @Failure 404 {object} api.Response "质押不存在"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /stakes/{id} [get]
func (h *StakeHandler) GetStake(c *gin.Context) {
	stakeID := c.Param("id")
	if stakeID == "" {
		h.errorResponse(c, http.StatusBadRequest, "stake ID is required")
		return
	}

	stake, err := h.stakeService.GetStake(c.Request.Context(), stakeID)
	if err != nil {
		h.logger.Error("Failed to get stake", zap.Error(err), zap.String("stake_id", stakeID))
		h.errorResponse(c, http.StatusNotFound, "Stake not found")
		return
	}

	h.successResponse(c, stake)
}

// GetUserStakes 获取用户质押列表
// @Summary 获取用户质押列表
// @Description 获取指定用户的质押列表
// @Tags 质押
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} api.Response{data=service.ListStakesResponse} "成功"
// @Failure 400 {object} api.Response "请求参数错误"
// @Failure 401 {object} api.Response "未授权"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /stakes/user/{user_id} [get]
func (h *StakeHandler) GetUserStakes(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.errorResponse(c, http.StatusBadRequest, "user ID is required")
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

	response, err := h.stakeService.GetUserStakes(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user stakes", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get user stakes")
		return
	}

	h.successResponse(c, response)
}

// RequestUnstake 请求解质押
// @Summary 请求解质押
// @Description 请求解质押指定的质押记录
// @Tags 质押
// @Security BearerAuth
// @Param id path string true "质押ID"
// @Success 200 {object} api.Response{data=stake.Stake} "成功"
// @Failure 400 {object} api.Response "请求参数错误"
// @Failure 401 {object} api.Response "未授权"
// @Failure 404 {object} api.Response "质押不存在"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /stakes/{id}/unstake [post]
func (h *StakeHandler) RequestUnstake(c *gin.Context) {
	userID := h.getUserID(c)
	stakeID := c.Param("id")

	if stakeID == "" {
		h.errorResponse(c, http.StatusBadRequest, "stake ID is required")
		return
	}

	stake, err := h.stakeService.RequestUnstake(c.Request.Context(), userID, stakeID)
	if err != nil {
		h.logger.Error("Failed to request unstake", zap.Error(err), zap.String("user_id", userID), zap.String("stake_id", stakeID))
		h.errorResponse(c, http.StatusBadRequest, "Failed to request unstake")
		return
	}

	h.successResponse(c, stake)
}

// CompleteUnstake 完成解质押
// @Summary 完成解质押
// @Description 完成解质押指定的质押记录
// @Tags 质押
// @Security BearerAuth
// @Param id path string true "质押ID"
// @Success 200 {object} api.Response{data=stake.Stake} "成功"
// @Failure 400 {object} api.Response "请求参数错误"
// @Failure 401 {object} api.Response "未授权"
// @Failure 404 {object} api.Response "质押不存在"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /stakes/{id}/complete [post]
func (h *StakeHandler) CompleteUnstake(c *gin.Context) {
	userID := h.getUserID(c)
	stakeID := c.Param("id")

	if stakeID == "" {
		h.errorResponse(c, http.StatusBadRequest, "stake ID is required")
		return
	}

	stake, err := h.stakeService.CompleteUnstake(c.Request.Context(), userID, stakeID)
	if err != nil {
		h.logger.Error("Failed to complete unstake", zap.Error(err), zap.String("user_id", userID), zap.String("stake_id", stakeID))
		h.errorResponse(c, http.StatusBadRequest, "Failed to complete unstake")
		return
	}

	h.successResponse(c, stake)
}

// ClaimRewards 领取奖励
// @Summary 领取奖励
// @Description 领取用户的质押奖励
// @Tags 质押
// @Security BearerAuth
// @Success 200 {object} api.Response{data=service.ClaimRewardsResponse} "成功"
// @Failure 401 {object} api.Response "未授权"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /stakes/rewards/claim [post]
func (h *StakeHandler) ClaimRewards(c *gin.Context) {
	userID := h.getUserID(c)

	response, err := h.stakeService.ClaimRewards(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to claim rewards", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to claim rewards")
		return
	}

	h.successResponse(c, response)
}

// GetUserStakeStats 获取用户质押统计信息
// @Summary 获取用户质押统计信息
// @Description 获取指定用户的质押统计信息
// @Tags 质押
// @Security BearerAuth
// @Param user_id path string true "用户ID"
// @Success 200 {object} api.Response{data=repository.UserStakeStats} "成功"
// @Failure 400 {object} api.Response "请求参数错误"
// @Failure 401 {object} api.Response "未授权"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /stakes/user/{user_id}/stats [get]
func (h *StakeHandler) GetUserStakeStats(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.errorResponse(c, http.StatusBadRequest, "user ID is required")
		return
	}

	stats, err := h.stakeService.GetUserStakeStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user stake stats", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get user stake stats")
		return
	}

	h.successResponse(c, stats)
}

// 管理员功能

// DistributeRewards 分发奖励（管理员功能）
// @Summary 分发奖励
// @Description 分发质押奖励给所有符合条件的用户（管理员功能）
// @Tags 管理员
// @Security BearerAuth
// @Success 200 {object} api.Response{data=service.DistributeRewardsResponse} "成功"
// @Failure 401 {object} api.Response "未授权"
// @Failure 403 {object} api.Response "权限不足"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /admin/stakes/distribute [post]
func (h *StakeHandler) DistributeRewards(c *gin.Context) {
	response, err := h.stakeService.DistributeRewards(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to distribute rewards", zap.Error(err))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to distribute rewards")
		return
	}

	h.successResponse(c, response)
}

// GetStakeStats 获取质押统计信息（管理员功能）
// @Summary 获取质押统计信息
// @Description 获取系统质押统计信息（管理员功能）
// @Tags 管理员
// @Security BearerAuth
// @Success 200 {object} api.Response{data=repository.StakeStats} "成功"
// @Failure 401 {object} api.Response "未授权"
// @Failure 403 {object} api.Response "权限不足"
// @Failure 500 {object} api.Response "服务器错误"
// @Router /admin/stakes/stats [get]
func (h *StakeHandler) GetStakeStats(c *gin.Context) {
	stats, err := h.stakeService.GetStakeStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get stake stats", zap.Error(err))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get stake stats")
		return
	}

	h.successResponse(c, stats)
}

// 辅助函数
func (h *StakeHandler) getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
