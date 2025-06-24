package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/logger"
	"bossfi-blockchain-backend/pkg/mreturn"
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

// GenerateNonce 生成登录nonce
// @Summary 生成登录nonce
// @Tags auth
// @Accept json
// @Produce json
// @Router /v1/auth/nonce [post]
func (h *UserHandler) GenerateNonce(c *gin.Context) {
	var req NonceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid nonce request", zap.Error(err))
		mreturn.BadRequest(c, "Invalid request parameters")
		return
	}

	nonce, err := h.userService.GenerateNonce(c.Request.Context(), req.WalletAddress)
	if err != nil {
		h.logger.Error("Failed to generate nonce", zap.Error(err), zap.String("wallet_address", req.WalletAddress))
		mreturn.InternalServerError(c, "Failed to generate nonce")
		return
	}

	mreturn.Success(c, NonceResponse{Message: nonce})
}

// Login 用户登录
// @Summary 用户登录
// @Tags auth
// @Accept json
// @Produce json
// @Router /v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		mreturn.BadRequest(c, "Invalid request parameters")
		return
	}

	serviceReq := &service.LoginRequest{
		WalletAddress: req.WalletAddress,
		Signature:     req.Signature,
		Message:       req.Message,
	}

	loginResponse, err := h.userService.Login(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err), zap.String("wallet_address", req.WalletAddress))
		mreturn.Unauthorized(c, "Login failed")
		return
	}

	mreturn.Success(c, LoginResponse{
		Token: loginResponse.Token,
		User:  loginResponse.User,
	})
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Tags 用户
// @Security BearerAuth
// @Produce json
// @Router /v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err), zap.String("user_id", userID))
		mreturn.NotFound(c, "User not found")
		return
	}

	mreturn.Success(c, user)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Tags 用户
// @Security BearerAuth
// @Accept json
// @Produce json
// @Router /v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserID(c)

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update profile request", zap.Error(err), zap.String("user_id", userID))
		mreturn.BadRequest(c, "Invalid request parameters")
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to update user profile", zap.Error(err), zap.String("user_id", userID))
		mreturn.Error(c, http.StatusConflict, "Failed to update profile")
		return
	}

	mreturn.Success(c, user)
}

// GetUserStats 获取用户统计信息
// @Summary 获取用户统计信息
// @Tags 用户
// @Security BearerAuth
// @Produce json
// @Router /v1/users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID := getUserID(c)

	stats, err := h.userService.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user stats", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to get user stats")
		return
	}

	mreturn.Success(c, stats)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Tags 用户
// @Produce json
// @Router /v1/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		mreturn.BadRequest(c, "keyword is required")
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

	search_response, err := h.userService.SearchUsers(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search users", zap.Error(err), zap.String("keyword", keyword))
		mreturn.InternalServerError(c, "Failed to search users")
		return
	}

	mreturn.Success(c, search_response)
}

// 管理员功能

// ListUsers 获取用户列表（管理员）
// @Summary 获取用户列表
// @Tags 管理员
// @Security BearerAuth
// @Produce json
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

	list_response, err := h.userService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		mreturn.InternalServerError(c, "Failed to list users")
		return
	}

	mreturn.Success(c, list_response)
}

// GetUserByID 根据ID获取用户（管理员）
// @Summary 根据ID获取用户
// @Tags 管理员
// @Security BearerAuth
// @Produce json
// @Router /v1/admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		mreturn.BadRequest(c, "user ID is required")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", userID))
		mreturn.NotFound(c, "User not found")
		return
	}

	mreturn.Success(c, user)
}

// DeleteUser 删除用户（管理员）
// @Summary 删除用户
// @Tags 管理员
// @Security BearerAuth
// @Router /v1/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		mreturn.BadRequest(c, "user ID is required")
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to delete user")
		return
	}

	mreturn.SuccessWithMessage(c, "User deleted successfully", nil)
}
