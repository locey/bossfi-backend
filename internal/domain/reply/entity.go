package reply

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostReply 帖子回复实体
type PostReply struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	PostID    string         `gorm:"type:char(36);not null;index" json:"post_id"`
	UserID    string         `gorm:"type:char(36);not null;index" json:"user_id"`
	ParentID  *string        `gorm:"type:char(36);index" json:"parent_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	LikeCount int64          `gorm:"default:0" json:"like_count"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM钩子
func (r *PostReply) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (PostReply) TableName() string {
	return "post_replies"
}

// IsTopLevel 检查是否是顶级回复
func (r *PostReply) IsTopLevel() bool {
	return r.ParentID == nil
}

// IsSubReply 检查是否是子回复
func (r *PostReply) IsSubReply() bool {
	return r.ParentID != nil
}

// IncrementLike 增加点赞数
func (r *PostReply) IncrementLike() {
	r.LikeCount++
}

// DecrementLike 减少点赞数
func (r *PostReply) DecrementLike() {
	if r.LikeCount > 0 {
		r.LikeCount--
	}
}
