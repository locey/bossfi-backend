package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/logger"
)

// UserHandler 用户处理
type UserHandler struct {
	userService service.UserService
	validator   *validator.Validate
	logger      *logger.Logger
}

// NewUserHandler 创建用户处理
func NewUserHandler(userService service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
		logger:      logger,
	}
}

// 请求和响应结
type NonceRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type NonceResponse struct {
	Message string `json:"message"`
}

type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Message       string `json:"message" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// 响应辅助函数
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func errorResponse(c *gin.Context, status int, message string, err error) {
	response := gin.H{
		"code":    status,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}
	c.JSON(status, response)
}

// GenerateNonce 生成登录nonce
// @Summary 生成登录nonce
// @Description 为钱包地址生成登录nonce
// @Tags auth
// @Accept json
// @Produce json
// @Param request body NonceRequest true "nonce请求"
// @Success 200 {object} NonceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/auth/nonce [post]
func (h *UserHandler) GenerateNonce(c *gin.Context) {
	var req NonceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid nonce request", zap.Error(err))
		errorResponse(c, http.StatusBadRequest, "Invalid request parameters", err)
		return
	}

	nonce, err := h.userService.GenerateNonce(c.Request.Context(), req.WalletAddress)
	if err != nil {
		h.logger.Error("Failed to generate nonce", zap.Error(err), zap.String("wallet_address", req.WalletAddress))
		errorResponse(c, http.StatusInternalServerError, "Failed to generate nonce", err)
		return
	}

	successResponse(c, NonceResponse{Message: nonce})
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用钱包签名进行用户登录
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		errorResponse(c, http.StatusBadRequest, "Invalid request parameters", err)
		return
	}

	serviceReq := &service.LoginRequest{
		WalletAddress: req.WalletAddress,
		Signature:     req.Signature,
		Message:       req.Message,
	}

	response, err := h.userService.Login(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err), zap.String("wallet_address", req.WalletAddress))
		errorResponse(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	successResponse(c, LoginResponse{
		Token: response.Token,
		User:  response.User,
	})
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前用户的详细资料信息
// @Tags 用户
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	successResponse(c, user)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前用户的资料信息
// @Tags 用户
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body service.UpdateProfileRequest true "更新资料请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserID(c)

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update profile request", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusBadRequest, "Invalid request parameters", err)
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to update user profile", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusConflict, "Failed to update profile", err)
		return
	}

	successResponse(c, user)
}

// GetUserStats 获取用户统计信息
// @Summary 获取用户统计信息
// @Description 获取当前用户的统计信息，包括帖子、质押等数据
// @Tags 用户
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID := getUserID(c)

	stats, err := h.userService.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user stats", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusInternalServerError, "Failed to get user stats", err)
		return
	}

	successResponse(c, stats)
}

// GetBalance 获取用户余额
// @Summary 获取用户余额
// @Description 获取当前用户的BOSS币、质押和奖励余额
// @Tags 用户
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/users/balance [get]
func (h *UserHandler) GetBalance(c *gin.Context) {
	userID := getUserID(c)

	balance, err := h.userService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusInternalServerError, "Failed to get balance", err)
		return
	}

	successResponse(c, balance)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户
// @Tags 用户
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		errorResponse(c, http.StatusBadRequest, "keyword is required", nil)
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

	response, err := h.userService.SearchUsers(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search users", zap.Error(err), zap.String("keyword", keyword))
		errorResponse(c, http.StatusInternalServerError, "Failed to search users", err)
		return
	}

	successResponse(c, response)
}

// 管理员功能

// ListUsers 获取用户列表（管理员）
// @Summary 获取用户列表
// @Description 获取所有用户列表（管理员功能）
// @Tags 管理员
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.userService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		errorResponse(c, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	successResponse(c, response)
}

// GetUserByID 根据ID获取用户（管理员）
// @Summary 根据ID获取用户
// @Description 根据用户ID获取用户详细信息（管理员功能）
// @Tags 管理员
// @Security BearerAuth
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	successResponse(c, user)
}

// DeleteUser 删除用户（管理员）
// @Summary 删除用户
// @Description 删除指定用户（管理员功能）
// @Tags 管理员
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", userID))
		errorResponse(c, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	successResponse(c, gin.H{"message": "User deleted successfully"})
}
