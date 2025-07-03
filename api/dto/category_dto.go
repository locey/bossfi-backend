package dto

import (
	"time"
)

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,max=50" example:"技术"`
	Description string `json:"description" binding:"max=200" example:"技术相关文章"`
	Icon        string `json:"icon" binding:"max=100" example:"tech-icon"`
	Color       string `json:"color" binding:"max=7" example:"#FF5733"`
	SortOrder   int    `json:"sort_order" example:"1"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required,max=50" example:"技术"`
	Description string `json:"description" binding:"max=200" example:"技术相关文章"`
	Icon        string `json:"icon" binding:"max=100" example:"tech-icon"`
	Color       string `json:"color" binding:"max=7" example:"#FF5733"`
	SortOrder   int    `json:"sort_order" example:"1"`
	IsActive    bool   `json:"is_active" example:"true"`
}

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID           uint      `json:"id" example:"1"`
	Name         string    `json:"name" example:"技术"`
	Description  string    `json:"description" example:"技术相关文章"`
	Icon         string    `json:"icon" example:"tech-icon"`
	Color        string    `json:"color" example:"#FF5733"`
	SortOrder    int       `json:"sort_order" example:"1"`
	IsActive     bool      `json:"is_active" example:"true"`
	ArticleCount int64     `json:"article_count" example:"10"`
	CreatedAt    time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CategoryListResponse 分类列表响应
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int64              `json:"total" example:"10"`
}

// CategoryQueryRequest 分类查询请求
type CategoryQueryRequest struct {
	Page     int   `form:"page" binding:"min=1" example:"1"`
	PageSize int   `form:"page_size" binding:"min=1,max=50" example:"10"`
	IsActive *bool `form:"is_active" example:"true"` // 是否只查询活跃分类
}
