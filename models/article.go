package models

import (
	"time"
)

type Article struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	CategoryID   *uint     `json:"category_id" gorm:"index"` // 分类ID，可为空
	Title        string    `json:"title" gorm:"type:varchar(200);not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	Images       []string  `json:"images" gorm:"type:jsonb;serializer:json"`
	LikeCount    int       `json:"like_count" gorm:"default:0"`
	CommentCount int       `json:"comment_count" gorm:"default:0"`
	ViewCount    int       `json:"view_count" gorm:"default:0"`
	IsDeleted    bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null"`

	// 关联关系
	User     User             `json:"user" gorm:"foreignKey:UserID"`
	Category *ArticleCategory `json:"category" gorm:"foreignKey:CategoryID"`
	Likes    []ArticleLike    `json:"likes" gorm:"foreignKey:ArticleID"`
	Comments []ArticleComment `json:"comments" gorm:"foreignKey:ArticleID"`
}

func (a *Article) TableName() string {
	return "articles"
}
