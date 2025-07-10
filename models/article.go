package models

import (
	"time"
)

type Article struct {
	ID           uint     `json:"id" gorm:"primarykey"`
	UserID       uint     `json:"user_id" gorm:"not null"`
	CategoryID   *uint    `json:"category_id" gorm:"index"` // 分类ID，可为空
	Title        string   `json:"title" gorm:"type:varchar(200);not null"`
	Content      string   `json:"content" gorm:"type:text;not null"`
	Images       []string `json:"images" gorm:"type:jsonb;serializer:json"`
	LikeCount    int      `json:"like_count" gorm:"default:0"`
	CommentCount int      `json:"comment_count" gorm:"default:0"`
	ViewCount    int      `json:"view_count" gorm:"default:0"`
	IsDeleted    bool     `json:"is_deleted" gorm:"default:false"`
	// AI评分相关字段
	Score       *float64   `json:"score" gorm:"type:decimal(3,2)"`   // AI评分 (0-10)
	ScoreTime   *time.Time `json:"score_time" gorm:"type:timestamp"` // AI评分时间
	ScoreReason string     `json:"score_reason" gorm:"type:text"`    // AI评分理由
	ScoreStatus int        `json:"score_status" gorm:"default:0"`    // 评分状态: 0-待评分, 1-评分中, 2-评分成功, -1-评分失败
	CreatedAt   time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"not null"`

	// 关联关系
	User     User             `json:"user" gorm:"foreignKey:UserID"`
	Category *ArticleCategory `json:"category" gorm:"foreignKey:CategoryID"`
	Likes    []ArticleLike    `json:"likes" gorm:"foreignKey:ArticleID"`
	Comments []ArticleComment `json:"comments" gorm:"foreignKey:ArticleID"`
}

func (a *Article) TableName() string {
	return "articles"
}
