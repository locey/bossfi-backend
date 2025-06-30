package models

import (
	"time"
)

type ArticleLike struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	ArticleID uint      `json:"article_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	// 关联关系
	User    User    `json:"user" gorm:"foreignKey:UserID"`
	Article Article `json:"article" gorm:"foreignKey:ArticleID"`
}

func (al *ArticleLike) TableName() string {
	return "article_likes"
}
