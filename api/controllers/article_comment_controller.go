package controllers

import (
	"net/http"
	"strconv"

	"bossfi-backend/api/dto"
	"bossfi-backend/app/services"
	"bossfi-backend/middleware"
	"bossfi-backend/models"

	"github.com/gin-gonic/gin"
)

type ArticleCommentController struct {
	commentService *services.ArticleCommentService
}

func NewArticleCommentController() *ArticleCommentController {
	return &ArticleCommentController{
		commentService: services.NewArticleCommentService(),
	}
}

// CreateComment 创建评论
// @Summary 创建评论
// @Description 为文章创建新评论，支持回复其他评论
// @Tags 评论
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateCommentRequest true "创建评论信息"
// @Success 200 {object} dto.CommentResponse "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /comments [post]
func (cc *ArticleCommentController) CreateComment(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	comment, err := cc.commentService.CreateComment(uint(userIDUint), req.ArticleID, req.ParentID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comment)
}

// GetComments 获取评论列表
// @Summary 获取评论列表
// @Description 分页获取文章评论列表，支持按父评论筛选
// @Tags 评论
// @Accept json
// @Produce json
// @Param article_id query int true "文章ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param parent_id query int false "父评论ID，用于获取回复"
// @Success 200 {object} dto.CommentListResponse "评论列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /comments [get]
func (cc *ArticleCommentController) GetComments(c *gin.Context) {
	var req dto.CommentQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 获取评论列表
	comments, _, err := cc.commentService.GetComments(req.ArticleID, req.Page, req.PageSize, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

// LikeComment 点赞评论
// @Summary 点赞评论
// @Description 为指定评论点赞或取消点赞
// @Tags 评论
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "评论ID"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /comments/{id}/like [post]
func (cc *ArticleCommentController) LikeComment(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	err = cc.commentService.LikeComment(uint(commentID), uint(userIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// GetUserComments 获取登录用户的所有评论
// @Summary 获取登录用户的所有评论
// @Description 分页获取当前登录用户的所有评论，包含文章信息和父评论信息
// @Tags 评论
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} dto.UserCommentListResponse "用户评论列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /user/comments [get]
func (cc *ArticleCommentController) GetUserComments(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UserCommentQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// 获取用户评论列表
	comments, total, err := cc.commentService.GetUserComments(uint(userIDUint), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	response := dto.UserCommentListResponse{
		Comments: make([]dto.UserCommentResponse, len(comments)),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	for i, comment := range comments {
		response.Comments[i] = cc.convertToUserCommentResponse(&comment)
	}

	c.JSON(http.StatusOK, response)
}

// convertToUserCommentResponse 转换为用户评论响应格式
func (cc *ArticleCommentController) convertToUserCommentResponse(comment *models.ArticleComment) dto.UserCommentResponse {
	response := dto.UserCommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		ArticleID: comment.ArticleID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		LikeCount: comment.LikeCount,
		IsDeleted: comment.IsDeleted,
		CreatedAt: comment.CreatedAt,
		User: dto.UserInfo{
			ID:            comment.User.ID,
			Username:      comment.User.Username,
			Avatar:        comment.User.Avatar,
			WalletAddress: comment.User.WalletAddress,
		},
	}

	// 添加文章信息
	if comment.Article.ID != 0 {
		// 截取文章内容前100字符
		content := comment.Article.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		response.Article = dto.ArticleInfo{
			ID:           comment.Article.ID,
			Title:        comment.Article.Title,
			Content:      content,
			CategoryID:   comment.Article.CategoryID,
			LikeCount:    comment.Article.LikeCount,
			CommentCount: comment.Article.CommentCount,
			ViewCount:    comment.Article.ViewCount,
			CreatedAt:    comment.Article.CreatedAt,
			User: dto.UserInfo{
				ID:            comment.Article.User.ID,
				Username:      comment.Article.User.Username,
				Avatar:        comment.Article.User.Avatar,
				WalletAddress: comment.Article.User.WalletAddress,
			},
		}

		// 添加分类信息
		if comment.Article.Category != nil {
			response.Article.Category = &dto.CategoryInfo{
				ID:          comment.Article.Category.ID,
				Name:        comment.Article.Category.Name,
				Description: comment.Article.Category.Description,
				Icon:        comment.Article.Category.Icon,
				Color:       comment.Article.Category.Color,
			}
		}
	}

	// 添加父评论信息（如果存在）
	if comment.Parent != nil {
		// 截取父评论内容前50字符
		parentContent := comment.Parent.Content
		if len(parentContent) > 50 {
			parentContent = parentContent[:50] + "..."
		}

		response.Parent = &dto.CommentInfo{
			ID:        comment.Parent.ID,
			Content:   parentContent,
			LikeCount: comment.Parent.LikeCount,
			CreatedAt: comment.Parent.CreatedAt,
			User: dto.UserInfo{
				ID:            comment.Parent.User.ID,
				Username:      comment.Parent.User.Username,
				Avatar:        comment.Parent.User.Avatar,
				WalletAddress: comment.Parent.User.WalletAddress,
			},
		}
	}

	return response
}
