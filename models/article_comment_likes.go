package models

import (
	"time"
)

type ArticleCommentLike struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	CommentID uint      `json:"comment_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	// 关联关系
	User    User           `json:"user" gorm:"foreignKey:UserID"`
	Comment ArticleComment `json:"comment" gorm:"foreignKey:CommentID"`
}

func (cl *ArticleCommentLike) TableName() string {
	return "article_comment_likes"
}
