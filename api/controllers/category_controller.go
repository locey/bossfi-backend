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

type CategoryController struct {
	categoryService *services.CategoryService
}

func NewCategoryController() *CategoryController {
	return &CategoryController{
		categoryService: services.NewCategoryService(),
	}
}

// CreateCategory 创建分类
// @Summary 创建分类
// @Description 创建新的文章分类
// @Tags categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateCategoryRequest true "创建分类信息"
// @Success 200 {object} dto.CategoryResponse "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /categories [post]
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	// 检查用户权限（这里可以添加管理员权限检查）
	_, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := cc.categoryService.CreateCategory(req.Name, req.Description, req.Icon, req.Color, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := cc.convertToCategoryResponse(category)
	c.JSON(http.StatusOK, response)
}

// GetCategory 获取分类详情
// @Summary 获取分类详情
// @Description 根据分类ID获取分类详细信息
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} dto.CategoryResponse "分类信息"
// @Failure 404 {object} map[string]interface{} "分类不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /categories/{id} [get]
func (cc *CategoryController) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	category, err := cc.categoryService.GetCategoryByID(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := cc.convertToCategoryResponse(category)
	c.JSON(http.StatusOK, response)
}

// GetCategories 获取分类列表
// @Summary 获取分类列表
// @Description 分页获取分类列表
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param is_active query bool false "是否只查询活跃分类"
// @Success 200 {object} dto.CategoryListResponse "分类列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /categories [get]
func (cc *CategoryController) GetCategories(c *gin.Context) {
	var req dto.CategoryQueryRequest
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

	categories, total, err := cc.categoryService.GetCategories(req.Page, req.PageSize, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	var responses []dto.CategoryResponse
	for _, category := range categories {
		responses = append(responses, *cc.convertToCategoryResponse(&category))
	}

	response := dto.CategoryListResponse{
		Categories: responses,
		Total:      total,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCategory 更新分类
// @Summary 更新分类
// @Description 更新分类信息
// @Tags categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "分类ID"
// @Param request body dto.UpdateCategoryRequest true "更新分类信息"
// @Success 200 {object} dto.CategoryResponse "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "分类不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /categories/{id} [put]
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	// 检查用户权限
	_, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := cc.categoryService.UpdateCategory(uint(id), req.Name, req.Description, req.Icon, req.Color, req.SortOrder, req.IsActive)
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := cc.convertToCategoryResponse(category)
	c.JSON(http.StatusOK, response)
}

// DeleteCategory 删除分类
// @Summary 删除分类
// @Description 删除分类（只能删除没有文章的分类）
// @Tags categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "分类ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "分类不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /categories/{id} [delete]
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	// 检查用户权限
	_, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	err = cc.categoryService.DeleteCategory(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "cannot delete category with existing articles" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

// GetAllActiveCategories 获取所有活跃分类
// @Summary 获取所有活跃分类
// @Description 获取所有活跃的分类列表（用于前端下拉选择）
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} dto.CategoryListResponse "活跃分类列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /categories/active [get]
func (cc *CategoryController) GetAllActiveCategories(c *gin.Context) {
	categories, err := cc.categoryService.GetAllActiveCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	var responses []dto.CategoryResponse
	for _, category := range categories {
		responses = append(responses, *cc.convertToCategoryResponse(&category))
	}

	response := dto.CategoryListResponse{
		Categories: responses,
		Total:      int64(len(responses)),
	}

	c.JSON(http.StatusOK, response)
}

// convertToCategoryResponse 转换为分类响应格式
func (cc *CategoryController) convertToCategoryResponse(category *models.ArticleCategory) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:           category.ID,
		Name:         category.Name,
		Description:  category.Description,
		Icon:         category.Icon,
		Color:        category.Color,
		SortOrder:    category.SortOrder,
		IsActive:     category.IsActive,
		ArticleCount: category.ArticleCount,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
	}
}
