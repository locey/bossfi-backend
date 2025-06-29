package controllers

import (
	"net/http"

	"bossfi-backend/app/services"
	"bossfi-backend/middleware"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	userService *services.UserService
}

func NewAuthController() *AuthController {
	return &AuthController{
		userService: services.NewUserService(),
	}
}

// GetNonceRequest 获取 nonce 请求结构
type GetNonceRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required" example:"0x1234567890123456789012345678901234567890"`
}

// GetNonceResponse 获取 nonce 响应结构
type GetNonceResponse struct {
	Message string `json:"message" example:"Welcome to BossFi!..."`
	Nonce   string `json:"nonce" example:"abc123def456"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required" example:"0x1234567890123456789012345678901234567890"`
	Signature     string `json:"signature" binding:"required" example:"0x1234567890abcdef..."`
	Message       string `json:"message" binding:"required" example:"Welcome to BossFi!..."`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  interface{} `json:"user"`
}

// GetNonce 获取签名消息和 nonce
// @Summary 获取用于钱包签名的消息和 nonce
// @Description 前端调用此接口获取需要签名的消息和 nonce，用于钱包登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body GetNonceRequest true "钱包地址信息"
// @Success 200 {object} GetNonceResponse "成功返回签名消息和nonce"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/nonce [post]
func (ac *AuthController) GetNonce(c *gin.Context) {
	logger := middleware.GetLoggerFromContext(c)
	var req GetNonceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request parameters")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters: " + err.Error(),
		})
		return
	}

	logger.WithField("wallet_address", req.WalletAddress).Info("Getting nonce for wallet")

	message, nonce, err := ac.userService.GetNonceMessage(req.WalletAddress)
	if err != nil {
		logger.WithError(err).Error("Failed to get nonce message")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.WithFields(map[string]interface{}{
		"wallet_address": req.WalletAddress,
		"nonce":          nonce,
	}).Info("Nonce generated successfully")

	c.JSON(http.StatusOK, GetNonceResponse{
		Message: message,
		Nonce:   nonce,
	})
}

// Login 钱包签名登录
// @Summary 钱包签名登录
// @Description 使用钱包签名进行登录验证，验证成功后返回JWT令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse "登录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "认证失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	logger := middleware.GetLoggerFromContext(c)
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid login request parameters")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters: " + err.Error(),
		})
		return
	}

	logger.WithField("wallet_address", req.WalletAddress).Info("Attempting wallet login")

	user, token, err := ac.userService.VerifyAndLogin(req.WalletAddress, req.Signature, req.Message)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"wallet_address": req.WalletAddress,
			"error":          err.Error(),
		}).Error("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 隐藏敏感信息
	userResponse := gin.H{
		"id":             user.ID,
		"wallet_address": user.WalletAddress,
		"username":       user.Username,
		"avatar":         user.Avatar,
		"boss_balance":   user.BossBalance,
		"staked_amount":  user.StakedAmount,
		"reward_balance": user.RewardBalance,
		"created_at":     user.CreatedAt,
		"last_login_at":  user.LastLoginAt,
	}

	logger.WithFields(map[string]interface{}{
		"user_id":        user.ID,
		"wallet_address": req.WalletAddress,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  userResponse,
	})
}

// GetProfile 获取用户个人信息
// @Summary 获取用户个人信息
// @Description 获取当前登录用户的详细个人信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "用户信息"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/profile [get]
func (ac *AuthController) GetProfile(c *gin.Context) {
	logger := middleware.GetLoggerFromContext(c)
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		logger.Error("User not authenticated - no user ID in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	logger.WithField("user_id", userID).Info("Getting user profile")

	user, err := ac.userService.GetUserByID(userID)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user profile")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// 隐藏敏感信息
	userResponse := gin.H{
		"id":                  user.ID,
		"wallet_address":      user.WalletAddress,
		"username":            user.Username,
		"email":               user.Email,
		"avatar":              user.Avatar,
		"bio":                 user.Bio,
		"boss_balance":        user.BossBalance,
		"staked_amount":       user.StakedAmount,
		"reward_balance":      user.RewardBalance,
		"is_profile_complete": user.IsProfileComplete,
		"created_at":          user.CreatedAt,
		"last_login_at":       user.LastLoginAt,
	}

	logger.WithField("user_id", userID).Info("User profile retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"user": userResponse,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，清除服务器端会话信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "登出成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	logger := middleware.GetLoggerFromContext(c)
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		logger.Error("User not authenticated - no user ID in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	logger.WithField("user_id", userID).Info("User logging out")

	if err := ac.userService.Logout(userID); err != nil {
		logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to logout user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
		})
		return
	}

	logger.WithField("user_id", userID).Info("User logged out successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
