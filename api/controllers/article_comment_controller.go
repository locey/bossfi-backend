package controllers

import (
	"net/http"
	"strconv"

	"bossfi-backend/api/dto"
	"bossfi-backend/app/services"
	"bossfi-backend/middleware"

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
