package models

import (
	"time"
)

// ArticleCategory 文章分类模型
type ArticleCategory struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	Name         string    `json:"name" gorm:"type:varchar(50);not null;unique"`
	Description  string    `json:"description" gorm:"type:varchar(200)"`
	Icon         string    `json:"icon" gorm:"type:varchar(100)"`
	Color        string    `json:"color" gorm:"type:varchar(7)"` // 十六进制颜色值，如 #FF5733
	SortOrder    int       `json:"sort_order" gorm:"default:0"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	ArticleCount int64     `json:"article_count" gorm:"-"` // 文章数量，不存储在数据库中
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`

	// 关联关系
	Articles []Article `json:"articles" gorm:"foreignKey:CategoryID"`
}

func (ac *ArticleCategory) TableName() string {
	return "article_categories"
}
