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

type ArticleController struct {
	articleService *services.ArticleService
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		articleService: services.NewArticleService(),
	}
}

// CreateArticle 创建文章
// @Summary 创建文章
// @Description 创建一篇新文章
// @Tags 文章
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateArticleRequest true "创建文章信息"
// @Success 200 {object} dto.ArticleResponse "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles [post]
func (ac *ArticleController) CreateArticle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	article, err := ac.articleService.CreateArticle(uint(userIDUint), req.Title, req.Content, req.Images, req.CategoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	response := ac.convertToArticleResponse(article)
	c.JSON(http.StatusOK, response)
}

// GetArticle 获取文章详情
// @Summary 获取文章详情
// @Description 根据文章ID获取文章详细信息
// @Tags 文章
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} dto.ArticleResponse "文章信息"
// @Failure 404 {object} map[string]interface{} "文章不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles/{id} [get]
func (ac *ArticleController) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	article, err := ac.articleService.GetArticleByID(uint(id))
	if err != nil {
		if err.Error() == "article not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := ac.convertToArticleResponse(article)
	c.JSON(http.StatusOK, response)
}

// GetArticles 获取文章列表
// @Summary 获取文章列表
// @Description 分页获取文章列表，支持排序、分类筛选和关键字搜索
// @Tags 文章
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param sort_by query string false "排序字段" Enums(created_at, like_count, view_count)
// @Param sort_order query string false "排序方向" Enums(asc, desc) default(desc)
// @Param user_id query int false "用户ID"
// @Param category_id query int false "分类ID"
// @Param keyword query string false "关键字搜索"
// @Success 200 {object} dto.ArticleListResponse "文章列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles [get]
func (ac *ArticleController) GetArticles(c *gin.Context) {
	var req dto.ArticleQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	articles, total, err := ac.articleService.GetArticles(req.Page, req.PageSize, req.SortBy, req.SortOrder, req.UserID, req.CategoryID, req.Keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	var responses []dto.ArticleResponse
	for _, article := range articles {
		responses = append(responses, *ac.convertToArticleResponse(&article))
	}

	response := dto.ArticleListResponse{
		Articles: responses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateArticle 更新文章
// @Summary 更新文章
// @Description 更新文章内容
// @Tags 文章
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "文章ID"
// @Param request body dto.UpdateArticleRequest true "更新文章信息"
// @Success 200 {object} dto.ArticleResponse "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "文章不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles/{id} [put]
func (ac *ArticleController) UpdateArticle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	var req dto.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	article, err := ac.articleService.UpdateArticle(uint(id), uint(userIDUint), req.Title, req.Content, req.Images, req.CategoryID)
	if err != nil {
		if err.Error() == "article not found or not authorized" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := ac.convertToArticleResponse(article)
	c.JSON(http.StatusOK, response)
}

// DeleteArticle 删除文章
// @Summary 删除文章
// @Description 删除文章（逻辑删除）
// @Tags 文章
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "文章不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles/{id} [delete]
func (ac *ArticleController) DeleteArticle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	err = ac.articleService.DeleteArticle(uint(id), uint(userIDUint))
	if err != nil {
		if err.Error() == "article not found or not authorized" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article deleted successfully"})
}

// LikeArticle 点赞文章
// @Summary 点赞文章
// @Description 给文章点赞
// @Tags 文章
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "点赞成功"
// @Failure 400 {object} map[string]interface{} "已经点赞"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "文章不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles/{id}/like [post]
func (ac *ArticleController) LikeArticle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	err = ac.articleService.LikeArticle(uint(id), uint(userIDUint))
	if err != nil {
		if err.Error() == "already liked" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if err.Error() == "article not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article liked successfully"})
}

// UnlikeArticle 取消点赞文章
// @Summary 取消点赞文章
// @Description 取消文章点赞
// @Tags 文章
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "取消点赞成功"
// @Failure 400 {object} map[string]interface{} "未点赞"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /articles/{id}/unlike [delete]
func (ac *ArticleController) UnlikeArticle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	err = ac.articleService.UnlikeArticle(uint(id), uint(userIDUint))
	if err != nil {
		if err.Error() == "like not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article unliked successfully"})
}

// convertToArticleResponse 转换为文章响应格式
func (ac *ArticleController) convertToArticleResponse(article *models.Article) *dto.ArticleResponse {
	response := &dto.ArticleResponse{
		ID:           article.ID,
		UserID:       article.UserID,
		CategoryID:   article.CategoryID,
		Title:        article.Title,
		Content:      article.Content,
		Images:       article.Images,
		LikeCount:    article.LikeCount,
		CommentCount: article.CommentCount,
		ViewCount:    article.ViewCount,
		IsDeleted:    article.IsDeleted,
		CreatedAt:    article.CreatedAt,
		UpdatedAt:    article.UpdatedAt,
	}

	// 添加用户信息
	if article.User.ID != 0 {
		response.User = dto.UserInfo{
			ID:            article.User.ID,
			Username:      article.User.Username,
			Avatar:        article.User.Avatar,
			WalletAddress: article.User.WalletAddress,
		}
	}

	// 添加分类信息
	if article.Category != nil && article.Category.ID != 0 {
		response.Category = &dto.CategoryInfo{
			ID:          article.Category.ID,
			Name:        article.Category.Name,
			Description: article.Category.Description,
			Icon:        article.Category.Icon,
			Color:       article.Category.Color,
		}
	}

	return response
}
