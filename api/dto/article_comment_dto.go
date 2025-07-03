package dto

import (
	"time"
)

// CreateCommentRequest 创建评论请求
// @Description 创建评论的请求参数
type CreateCommentRequest struct {
	ArticleID uint   `json:"article_id" binding:"required" example:"1"`            // 文章ID
	ParentID  *uint  `json:"parent_id" example:"1"`                                // 父评论ID，用于回复
	Content   string `json:"content" binding:"required,max=1000" example:"这是一条评论"` // 评论内容
}

// UpdateCommentRequest 更新评论请求
// @Description 更新评论的请求参数
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,max=1000" example:"更新后的评论内容"` // 评论内容
}

// CommentResponse 评论响应
// @Description 评论的响应数据结构
type CommentResponse struct {
	ID        uint      `json:"id" example:"1"`                            // 评论ID
	UserID    uint      `json:"user_id" example:"1"`                       // 用户ID
	ArticleID uint      `json:"article_id" example:"1"`                    // 文章ID
	ParentID  *uint     `json:"parent_id" example:"1"`                     // 父评论ID
	Content   string    `json:"content" example:"评论内容"`                    // 评论内容
	LikeCount int       `json:"like_count" example:"5"`                    // 点赞数
	IsDeleted bool      `json:"is_deleted" example:"false"`                // 是否已删除
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"` // 创建时间

	// 关联数据
	User    UserInfo          `json:"user"`    // 用户信息
	Replies []CommentResponse `json:"replies"` // 回复列表
}

// UserCommentResponse 用户评论响应（包含文章信息和父评论信息）
// @Description 用户评论的响应数据结构，包含完整的文章和父评论信息
type UserCommentResponse struct {
	ID        uint      `json:"id" example:"1"`                            // 评论ID
	UserID    uint      `json:"user_id" example:"1"`                       // 用户ID
	ArticleID uint      `json:"article_id" example:"1"`                    // 文章ID
	ParentID  *uint     `json:"parent_id" example:"1"`                     // 父评论ID
	Content   string    `json:"content" example:"评论内容"`                    // 评论内容
	LikeCount int       `json:"like_count" example:"5"`                    // 点赞数
	IsDeleted bool      `json:"is_deleted" example:"false"`                // 是否已删除
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"` // 创建时间

	// 关联数据
	User    UserInfo     `json:"user"`    // 用户信息
	Article ArticleInfo  `json:"article"` // 文章信息
	Parent  *CommentInfo `json:"parent"`  // 父评论信息（如果存在）
}

// ArticleInfo 文章信息（简化版）
// @Description 文章的基本信息
type ArticleInfo struct {
	ID           uint      `json:"id" example:"1"`                            // 文章ID
	Title        string    `json:"title" example:"文章标题"`                      // 文章标题
	Content      string    `json:"content" example:"文章内容"`                    // 文章内容（截取前100字符）
	CategoryID   *uint     `json:"category_id" example:"1"`                   // 分类ID
	LikeCount    int       `json:"like_count" example:"10"`                   // 点赞数
	CommentCount int       `json:"comment_count" example:"5"`                 // 评论数
	ViewCount    int       `json:"view_count" example:"100"`                  // 浏览数
	CreatedAt    time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"` // 创建时间

	// 关联数据
	User     UserInfo      `json:"user"`     // 文章作者信息
	Category *CategoryInfo `json:"category"` // 分类信息
}

// CommentInfo 评论信息（简化版）
// @Description 父评论的基本信息
type CommentInfo struct {
	ID        uint      `json:"id" example:"1"`                            // 评论ID
	Content   string    `json:"content" example:"父评论内容"`                   // 评论内容（截取前50字符）
	LikeCount int       `json:"like_count" example:"3"`                    // 点赞数
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"` // 创建时间

	// 关联数据
	User UserInfo `json:"user"` // 评论作者信息
}

// CommentListResponse 评论列表响应
// @Description 评论列表的响应数据结构
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`               // 评论列表
	Total    int64             `json:"total" example:"50"`     // 总评论数
	Page     int               `json:"page" example:"1"`       // 当前页码
	PageSize int               `json:"page_size" example:"10"` // 每页数量
}

// UserCommentListResponse 用户评论列表响应
// @Description 用户评论列表的响应数据结构
type UserCommentListResponse struct {
	Comments []UserCommentResponse `json:"comments"`               // 用户评论列表
	Total    int64                 `json:"total" example:"50"`     // 总评论数
	Page     int                   `json:"page" example:"1"`       // 当前页码
	PageSize int                   `json:"page_size" example:"10"` // 每页数量
}

// CommentQueryRequest 评论查询请求
// @Description 查询评论列表的请求参数
type CommentQueryRequest struct {
	ArticleID uint  `form:"article_id" binding:"required" example:"1"`     // 文章ID
	Page      int   `form:"page" binding:"min=1" example:"1"`              // 页码
	PageSize  int   `form:"page_size" binding:"min=1,max=50" example:"10"` // 每页数量
	ParentID  *uint `form:"parent_id" example:"1"`                         // 父评论ID，查询特定评论的回复
}

// UserCommentQueryRequest 用户评论查询请求
// @Description 查询用户评论列表的请求参数
type UserCommentQueryRequest struct {
	Page     int `form:"page" binding:"min=1" example:"1"`              // 页码
	PageSize int `form:"page_size" binding:"min=1,max=50" example:"10"` // 每页数量
}
