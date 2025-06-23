package post

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PostType 帖子类型
type PostType string

const (
	PostTypeJob        PostType = "job"        // 招聘
	PostTypeResume     PostType = "resume"     // 简历
	PostTypeDiscussion PostType = "discussion" // 讨论
)

// PostStatus 帖子状态
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"     // 草稿
	PostStatusPublished PostStatus = "published" // 已发布
	PostStatusClosed    PostStatus = "closed"    // 已关闭
)

// Tags 标签类型（JSON数组）
type Tags []string

// Scan 实现sql.Scanner接口
func (t *Tags) Scan(value interface{}) error {
	if value == nil {
		*t = Tags{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrInvalidTags
	}

	return json.Unmarshal(bytes, t)
}

// Value 实现driver.Valuer接口
func (t Tags) Value() (driver.Value, error) {
	if len(t) == 0 {
		return nil, nil
	}
	return json.Marshal(t)
}

// Post 帖子实体
type Post struct {
	ID           string          `gorm:"type:char(36);primaryKey" json:"id"`
	AuthorID     string          `gorm:"type:char(36);not null;index" json:"author_id"`
	TokenID      *string         `gorm:"type:varchar(100);uniqueIndex" json:"token_id"`
	Title        string          `gorm:"type:varchar(255);not null" json:"title"`
	Content      *string         `gorm:"type:longtext" json:"content"`
	PostType     PostType        `gorm:"type:enum('job','resume','discussion');not null" json:"post_type"`
	Status       PostStatus      `gorm:"type:enum('draft','published','closed');default:'draft'" json:"status"`
	Tags         Tags            `gorm:"type:json" json:"tags"`
	Salary       *string         `gorm:"type:varchar(100)" json:"salary"`
	Location     *string         `gorm:"type:varchar(100)" json:"location"`
	Company      *string         `gorm:"type:varchar(100)" json:"company"`
	Requirements *string         `gorm:"type:text" json:"requirements"`
	BossCost     decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"boss_cost"`
	ViewCount    int64           `gorm:"default:0" json:"view_count"`
	LikeCount    int64           `gorm:"default:0" json:"like_count"`
	ReplyCount   int64           `gorm:"default:0" json:"reply_count"`
	IPFSHash     *string         `gorm:"type:varchar(255)" json:"ipfs_hash"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

// BeforeCreate GORM钩子
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Post) TableName() string {
	return "posts"
}

// IsPublished 检查是否已发布
func (p *Post) IsPublished() bool {
	return p.Status == PostStatusPublished
}

// IsDraft 检查是否是草稿
func (p *Post) IsDraft() bool {
	return p.Status == PostStatusDraft
}

// IsClosed 检查是否已关闭
func (p *Post) IsClosed() bool {
	return p.Status == PostStatusClosed
}

// CanEdit 检查是否可以编辑
func (p *Post) CanEdit() bool {
	return p.Status == PostStatusDraft
}

// CanPublish 检查是否可以发布
func (p *Post) CanPublish() bool {
	return p.Status == PostStatusDraft && p.Title != "" && p.Content != nil && *p.Content != ""
}

// Publish 发布帖子
func (p *Post) Publish() error {
	if !p.CanPublish() {
		return ErrCannotPublish
	}
	p.Status = PostStatusPublished
	return nil
}

// Close 关闭帖子
func (p *Post) Close() error {
	if p.Status != PostStatusPublished {
		return ErrCannotClose
	}
	p.Status = PostStatusClosed
	return nil
}

// IncrementView 增加浏览次数
func (p *Post) IncrementView() {
	p.ViewCount++
}

// IncrementLike 增加点赞数
func (p *Post) IncrementLike() {
	p.LikeCount++
}

// DecrementLike 减少点赞数
func (p *Post) DecrementLike() {
	if p.LikeCount > 0 {
		p.LikeCount--
	}
}

// IncrementReply 增加回复数
func (p *Post) IncrementReply() {
	p.ReplyCount++
}

// DecrementReply 减少回复数
func (p *Post) DecrementReply() {
	if p.ReplyCount > 0 {
		p.ReplyCount--
	}
}

// AddTag 添加标签
func (p *Post) AddTag(tag string) {
	if p.Tags == nil {
		p.Tags = Tags{}
	}

	// 检查标签是否已存在
	for _, existingTag := range p.Tags {
		if existingTag == tag {
			return
		}
	}

	p.Tags = append(p.Tags, tag)
}

// RemoveTag 移除标签
func (p *Post) RemoveTag(tag string) {
	if p.Tags == nil {
		return
	}

	for i, existingTag := range p.Tags {
		if existingTag == tag {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			return
		}
	}
}

// HasTag 检查是否包含标签
func (p *Post) HasTag(tag string) bool {
	if p.Tags == nil {
		return false
	}

	for _, existingTag := range p.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}
