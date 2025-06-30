package models

import (
	"time"
)

type ArticleComment struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	ArticleID uint      `json:"article_id" gorm:"not null"`
	ParentID  *uint     `json:"parent_id"` // 父评论ID，用于回复功能
	Content   string    `json:"content" gorm:"type:text;not null"`
	LikeCount int       `json:"like_count" gorm:"default:0"`
	IsDeleted bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	// 关联关系
	User    User                 `json:"user" gorm:"foreignKey:UserID"`
	Article Article              `json:"article" gorm:"foreignKey:ArticleID"`
	Parent  *ArticleComment      `json:"parent" gorm:"foreignKey:ParentID"`
	Replies []ArticleComment     `json:"replies" gorm:"foreignKey:ParentID"`
	Likes   []ArticleCommentLike `json:"likes" gorm:"foreignKey:CommentID"`
}

func (c *ArticleComment) TableName() string {
	return "article_comments"
}
