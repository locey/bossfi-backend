package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/logger"
)

// PostHandler 帖子处理器
type PostHandler struct {
	postService service.PostService
	logger      *logger.Logger
}

// NewPostHandler 创建帖子处理器
func NewPostHandler(postService service.PostService, logger *logger.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		logger:      logger,
	}
}

// errorResponse 错误响应
func (h *PostHandler) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"code":    status,
		"message": message,
	})
}

// successResponse 成功响应
func (h *PostHandler) successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// successResponseWithStatus 带状态码的成功响应
func (h *PostHandler) successResponseWithStatus(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// CreatePost 创建帖子
// @Summary 创建帖子
// @Description 创建新的帖子
// @Tags 帖子
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body service.CreatePostRequest true "创建帖子请求"
// @Success 201 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID := getUserID(c)

	var req service.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create post request", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	post, err := h.postService.CreatePost(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to create post")
		return
	}

	h.successResponseWithStatus(c, http.StatusCreated, post)
}

// GetPost 获取帖子详情
// @Summary 获取帖子详情
// @Description 根据ID获取帖子详细信息
// @Tags 帖子
// @Produce json
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	// 增加浏览次数
	if err := h.postService.ViewPost(c.Request.Context(), postID); err != nil {
		h.logger.Warn("Failed to increment view count", zap.Error(err), zap.String("post_id", postID))
	}

	post, err := h.postService.GetPost(c.Request.Context(), postID)
	if err != nil {
		h.logger.Error("Failed to get post", zap.Error(err), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusNotFound, "Post not found")
		return
	}

	h.successResponse(c, post)
}

// UpdatePost 更新帖子
// @Summary 更新帖子
// @Description 更新帖子信息
// @Tags 帖子
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "帖子ID"
// @Param request body service.UpdatePostRequest true "更新帖子请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	var req service.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update post request", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), userID, postID, &req)
	if err != nil {
		h.logger.Error("Failed to update post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusForbidden, "Failed to update post")
		return
	}

	h.successResponse(c, post)
}

// DeletePost 删除帖子
// @Summary 删除帖子
// @Description 删除指定帖子
// @Tags 帖子
// @Security BearerAuth
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	err := h.postService.DeletePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to delete post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusForbidden, "Failed to delete post")
		return
	}

	h.successResponse(c, gin.H{"message": "Post deleted successfully"})
}

// PublishPost 发布帖子
// @Summary 发布帖子
// @Description 将草稿状态的帖子发布
// @Tags 帖子
// @Security BearerAuth
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id}/publish [post]
func (h *PostHandler) PublishPost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	post, err := h.postService.PublishPost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to publish post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusForbidden, "Failed to publish post")
		return
	}

	h.successResponse(c, post)
}

// ClosePost 关闭帖子
// @Summary 关闭帖子
// @Description 关闭指定帖子
// @Tags 帖子
// @Security BearerAuth
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id}/close [post]
func (h *PostHandler) ClosePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	post, err := h.postService.ClosePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to close post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusForbidden, "Failed to close post")
		return
	}

	h.successResponse(c, post)
}

// ListPosts 获取帖子列表
// @Summary 获取帖子列表
// @Description 获取帖子列表，支持分页和过滤
// @Tags 帖子
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param type query string false "帖子类型"
// @Param status query string false "帖子状态"
// @Param company query string false "公司名称"
// @Param location query string false "工作地点"
// @Param tags query string false "标签"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts [get]
func (h *PostHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &service.ListPostsRequest{
		Page:     page,
		PageSize: pageSize,
		OrderBy:  "created_at",
		OrderDir: "desc",
	}

	// 处理过滤参数
	if postType := c.Query("type"); postType != "" {
		// 这里需要转换字符串到PostType枚举
		// 暂时跳过类型转换，让service层处理
	}

	if status := c.Query("status"); status != "" {
		// 这里需要转换字符串到PostStatus枚举
		// 暂时跳过类型转换，让service层处理
	}

	if company := c.Query("company"); company != "" {
		req.Company = &company
	}

	if location := c.Query("location"); location != "" {
		req.Location = &location
	}

	if tags := c.Query("tags"); tags != "" {
		// 简单按逗号分割标签
		req.Tags = []string{tags}
	}

	response, err := h.postService.ListPosts(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list posts", zap.Error(err))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to list posts")
		return
	}

	h.successResponse(c, response)
}

// GetUserPosts 获取用户帖子
// @Summary 获取用户帖子
// @Description 获取指定用户的帖子列表
// @Tags 帖子
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/users/{user_id}/posts [get]
func (h *PostHandler) GetUserPosts(c *gin.Context) {
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

	response, err := h.postService.GetUserPosts(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user posts", zap.Error(err), zap.String("user_id", userID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get user posts")
		return
	}

	h.successResponse(c, response)
}

// SearchPosts 搜索帖子
// @Summary 搜索帖子
// @Description 根据关键词搜索帖子
// @Tags 帖子
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param type query string false "帖子类型"
// @Param location query string false "工作地点"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/search [get]
func (h *PostHandler) SearchPosts(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		h.errorResponse(c, http.StatusBadRequest, "keyword is required")
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

	req := &service.SearchPostsRequest{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
		OrderBy:  "created_at",
		OrderDir: "desc",
	}

	if postType := c.Query("type"); postType != "" {
		// 这里需要转换字符串到PostType枚举
		// 暂时跳过类型转换，让service层处理
	}

	if location := c.Query("location"); location != "" {
		req.Location = &location
	}

	response, err := h.postService.SearchPosts(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search posts", zap.Error(err), zap.String("keyword", keyword))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to search posts")
		return
	}

	h.successResponse(c, response)
}

// GetPopularPosts 获取热门帖子
// @Summary 获取热门帖子
// @Description 获取热门帖子列表
// @Tags 帖子
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/popular [get]
func (h *PostHandler) GetPopularPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.postService.GetPopularPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get popular posts", zap.Error(err))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get popular posts")
		return
	}

	h.successResponse(c, response)
}

// GetTrendingPosts 获取趋势帖子
// @Summary 获取趋势帖子
// @Description 获取趋势帖子列表
// @Tags 帖子
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/trending [get]
func (h *PostHandler) GetTrendingPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.postService.GetTrendingPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get trending posts", zap.Error(err))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to get trending posts")
		return
	}

	h.successResponse(c, response)
}

// LikePost 点赞帖子
// @Summary 点赞帖子
// @Description 对指定帖子进行点赞
// @Tags 帖子
// @Security BearerAuth
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id}/like [post]
func (h *PostHandler) LikePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	err := h.postService.LikePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to like post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to like post")
		return
	}

	h.successResponse(c, gin.H{"message": "Post liked successfully"})
}

// UnlikePost 取消点赞帖子
// @Summary 取消点赞帖子
// @Description 取消对指定帖子的点赞
// @Tags 帖子
// @Security BearerAuth
// @Param id path string true "帖子ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "帖子不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /v1/posts/{id}/unlike [post]
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		h.errorResponse(c, http.StatusBadRequest, "post ID is required")
		return
	}

	err := h.postService.UnlikePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to unlike post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		h.errorResponse(c, http.StatusInternalServerError, "Failed to unlike post")
		return
	}

	h.successResponse(c, gin.H{"message": "Post unliked successfully"})
}
