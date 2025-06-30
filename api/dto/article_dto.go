package dto

import (
	"time"
)

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title   string   `json:"title" binding:"required,max=200" example:"我的第一篇文章"`
	Content string   `json:"content" binding:"required" example:"这是文章内容..."`
	Images  []string `json:"images" example:"https://example.com/image1.jpg,https://example.com/image2.jpg"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title   string   `json:"title" binding:"required,max=200" example:"更新后的标题"`
	Content string   `json:"content" binding:"required" example:"更新后的内容..."`
	Images  []string `json:"images" example:"https://example.com/image2.jpg,https://example.com/image3.jpg"`
}

// ArticleResponse 文章响应
type ArticleResponse struct {
	ID           uint      `json:"id" example:"1"`
	UserID       uint      `json:"user_id" example:"1"`
	Title        string    `json:"title" example:"文章标题"`
	Content      string    `json:"content" example:"文章内容"`
	Images       []string  `json:"images" example:"https://example.com/image.jpg,https://example.com/image2.jpg"`
	LikeCount    int       `json:"like_count" example:"10"`
	CommentCount int       `json:"comment_count" example:"5"`
	ViewCount    int       `json:"view_count" example:"100"`
	IsDeleted    bool      `json:"is_deleted" example:"false"`
	CreatedAt    time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-01-01T00:00:00Z"`

	// 关联数据
	User UserInfo `json:"user"`
}

// UserInfo 用户信息（简化版）
type UserInfo struct {
	ID            uint    `json:"id" example:"1"`
	Username      *string `json:"username" example:"用户名"`
	Avatar        *string `json:"avatar" example:"https://example.com/avatar.jpg"`
	WalletAddress string  `json:"wallet_address" example:"0x1234..."`
}

// ArticleListResponse 文章列表响应
type ArticleListResponse struct {
	Articles []ArticleResponse `json:"articles"`
	Total    int64             `json:"total" example:"100"`
	Page     int               `json:"page" example:"1"`
	PageSize int               `json:"page_size" example:"10"`
}

// ArticleQueryRequest 文章查询请求
type ArticleQueryRequest struct {
	Page      int    `form:"page" binding:"min=1" example:"1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=50" example:"10"`
	SortBy    string `form:"sort_by" example:"created_at"` // created_at, like_count, view_count
	SortOrder string `form:"sort_order" example:"desc"`    // asc, desc
	UserID    *uint  `form:"user_id" example:"1"`
}
