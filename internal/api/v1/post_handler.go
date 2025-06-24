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

// CreatePost 创建帖子
// @Summary 创建帖子
// @Tags 帖子
// @Security BearerAuth
// @Accept json
// @Produce json
// @Router /v1/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID := getUserID(c)

	var req service.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create post request", zap.Error(err), zap.String("user_id", userID))
		mreturn.BadRequest(c, "Invalid request parameters")
		return
	}

	post, err := h.postService.CreatePost(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to create post")
		return
	}

	c.JSON(http.StatusCreated, mreturn.Response{
		Code:    0,
		Message: "success",
		Data:    post,
	})
}

// GetPost 获取帖子详情
// @Summary 获取帖子详情
// @Tags 帖子
// @Produce json
// @Router /v1/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	// 增加浏览次数
	if err := h.postService.ViewPost(c.Request.Context(), postID); err != nil {
		h.logger.Warn("Failed to increment view count", zap.Error(err), zap.String("post_id", postID))
	}

	post, err := h.postService.GetPost(c.Request.Context(), postID)
	if err != nil {
		h.logger.Error("Failed to get post", zap.Error(err), zap.String("post_id", postID))
		mreturn.NotFound(c, "Post not found")
		return
	}

	mreturn.Success(c, post)
}

// UpdatePost 更新帖子
// @Summary 更新帖子
// @Tags 帖子
// @Security BearerAuth
// @Accept json
// @Produce json
// @Router /v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	var req service.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update post request", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.BadRequest(c, "Invalid request parameters")
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), userID, postID, &req)
	if err != nil {
		h.logger.Error("Failed to update post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.Forbidden(c, "Failed to update post")
		return
	}

	mreturn.Success(c, post)
}

// DeletePost 删除帖子
// @Summary 删除帖子
// @Tags 帖子
// @Security BearerAuth
// @Router /v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	err := h.postService.DeletePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to delete post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.Forbidden(c, "Failed to delete post")
		return
	}

	mreturn.Success(c, gin.H{"message": "Post deleted successfully"})
}

// PublishPost 发布帖子
// @Summary 发布帖子
// @Tags 帖子
// @Security BearerAuth
// @Router /v1/posts/{id}/publish [post]
func (h *PostHandler) PublishPost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	post, err := h.postService.PublishPost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to publish post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.Forbidden(c, "Failed to publish post")
		return
	}

	mreturn.Success(c, post)
}

// ClosePost 关闭帖子
// @Summary 关闭帖子
// @Tags 帖子
// @Security BearerAuth
// @Router /v1/posts/{id}/close [post]
func (h *PostHandler) ClosePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	post, err := h.postService.ClosePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to close post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.Forbidden(c, "Failed to close post")
		return
	}

	mreturn.Success(c, post)
}

// ListPosts 获取帖子列表
// @Summary 获取帖子列表
// @Tags 帖子
// @Produce json
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

	list_response, err := h.postService.ListPosts(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list posts", zap.Error(err))
		mreturn.InternalServerError(c, "Failed to list posts")
		return
	}

	mreturn.Success(c, list_response)
}

// GetUserPosts 获取用户帖子
// @Summary 获取用户帖子
// @Tags 帖子
// @Produce json
// @Router /v1/users/{user_id}/posts [get]
func (h *PostHandler) GetUserPosts(c *gin.Context) {
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

	user_posts_response, err := h.postService.GetUserPosts(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user posts", zap.Error(err), zap.String("user_id", userID))
		mreturn.InternalServerError(c, "Failed to get user posts")
		return
	}

	mreturn.Success(c, user_posts_response)
}

// SearchPosts 搜索帖子
// @Summary 搜索帖子
// @Tags 帖子
// @Produce json
// @Router /v1/posts/search [get]
func (h *PostHandler) SearchPosts(c *gin.Context) {
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

	search_response, err := h.postService.SearchPosts(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search posts", zap.Error(err), zap.String("keyword", keyword))
		mreturn.InternalServerError(c, "Failed to search posts")
		return
	}

	mreturn.Success(c, search_response)
}

// GetPopularPosts 获取热门帖子
// @Summary 获取热门帖子
// @Tags 帖子
// @Produce json
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

	popular_response, err := h.postService.GetPopularPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get popular posts", zap.Error(err))
		mreturn.InternalServerError(c, "Failed to get popular posts")
		return
	}

	mreturn.Success(c, popular_response)
}

// GetTrendingPosts 获取趋势帖子
// @Summary 获取趋势帖子
// @Tags 帖子
// @Produce json
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

	trending_response, err := h.postService.GetTrendingPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get trending posts", zap.Error(err))
		mreturn.InternalServerError(c, "Failed to get trending posts")
		return
	}

	mreturn.Success(c, trending_response)
}

// LikePost 点赞帖子
// @Summary 点赞帖子
// @Tags 帖子
// @Security BearerAuth
// @Router /v1/posts/{id}/like [post]
func (h *PostHandler) LikePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	err := h.postService.LikePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to like post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.InternalServerError(c, "Failed to like post")
		return
	}

	mreturn.Success(c, gin.H{"message": "Post liked successfully"})
}

// UnlikePost 取消点赞帖子
// @Summary 取消点赞帖子
// @Tags 帖子
// @Security BearerAuth
// @Router /v1/posts/{id}/unlike [post]
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID := getUserID(c)
	postID := c.Param("id")

	if postID == "" {
		mreturn.BadRequest(c, "post ID is required")
		return
	}

	err := h.postService.UnlikePost(c.Request.Context(), userID, postID)
	if err != nil {
		h.logger.Error("Failed to unlike post", zap.Error(err), zap.String("user_id", userID), zap.String("post_id", postID))
		mreturn.InternalServerError(c, "Failed to unlike post")
		return
	}

	mreturn.Success(c, gin.H{"message": "Post unliked successfully"})
}
