package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"bossfi-blockchain-backend/internal/domain/post"
)

// PostRepository 帖子仓储接口
type PostRepository interface {
	Create(ctx context.Context, p *post.Post) error
	GetByID(ctx context.Context, id string) (*post.Post, error)
	GetByTokenID(ctx context.Context, tokenID string) (*post.Post, error)
	Update(ctx context.Context, p *post.Post) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter PostFilter) ([]*post.Post, int64, error)
	GetByAuthor(ctx context.Context, authorID string, offset, limit int) ([]*post.Post, int64, error)
	IncrementView(ctx context.Context, id string) error
	IncrementLike(ctx context.Context, id string) error
	DecrementLike(ctx context.Context, id string) error
	IncrementReply(ctx context.Context, id string) error
	DecrementReply(ctx context.Context, id string) error
	Search(ctx context.Context, keyword string, filter PostFilter) ([]*post.Post, int64, error)
	GetPopular(ctx context.Context, offset, limit int) ([]*post.Post, int64, error)
	GetTrending(ctx context.Context, offset, limit int) ([]*post.Post, int64, error)
}

// PostFilter 帖子过滤条件
type PostFilter struct {
	PostType *post.PostType   `json:"post_type"`
	Status   *post.PostStatus `json:"status"`
	AuthorID *string          `json:"author_id"`
	Tags     []string         `json:"tags"`
	Location *string          `json:"location"`
	Company  *string          `json:"company"`
	Offset   int              `json:"offset"`
	Limit    int              `json:"limit"`
	OrderBy  string           `json:"order_by"`  // created_at, view_count, like_count, reply_count
	OrderDir string           `json:"order_dir"` // asc, desc
}

// postRepository 帖子仓储实现
type postRepository struct {
	db *gorm.DB
}

// NewPostRepository 创建帖子仓储
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

// Create 创建帖子
func (r *postRepository) Create(ctx context.Context, p *post.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// GetByID 根据ID获取帖子
func (r *postRepository) GetByID(ctx context.Context, id string) (*post.Post, error) {
	var p post.Post
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, post.ErrPostNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetByTokenID 根据TokenID获取帖子
func (r *postRepository) GetByTokenID(ctx context.Context, tokenID string) (*post.Post, error) {
	var p post.Post
	err := r.db.WithContext(ctx).Where("token_id = ?", tokenID).First(&p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, post.ErrPostNotFound
		}
		return nil, err
	}
	return &p, nil
}

// Update 更新帖子
func (r *postRepository) Update(ctx context.Context, p *post.Post) error {
	return r.db.WithContext(ctx).Save(p).Error
}

// Delete 删除帖子
func (r *postRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&post.Post{}, "id = ?", id).Error
}

// List 获取帖子列表
func (r *postRepository) List(ctx context.Context, filter PostFilter) ([]*post.Post, int64, error) {
	var posts []*post.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&post.Post{})

	// 应用过滤条件
	query = r.applyFilter(query, filter)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序和分页
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "DESC"
	if filter.OrderDir != "" {
		orderDir = filter.OrderDir
	}

	err := query.
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order(fmt.Sprintf("%s %s", orderBy, orderDir)).
		Find(&posts).Error

	return posts, total, err
}

// GetByAuthor 获取作者的帖子
func (r *postRepository) GetByAuthor(ctx context.Context, authorID string, offset, limit int) ([]*post.Post, int64, error) {
	var posts []*post.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&post.Post{}).Where("author_id = ?", authorID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&posts).Error

	return posts, total, err
}

// IncrementView 增加浏览次数
func (r *postRepository) IncrementView(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&post.Post{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).
		Error
}

// IncrementLike 增加点赞数
func (r *postRepository) IncrementLike(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&post.Post{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).
		Error
}

// DecrementLike 减少点赞数
func (r *postRepository) DecrementLike(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&post.Post{}).
		Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).
		Error
}

// IncrementReply 增加回复数
func (r *postRepository) IncrementReply(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&post.Post{}).
		Where("id = ?", id).
		UpdateColumn("reply_count", gorm.Expr("reply_count + ?", 1)).
		Error
}

// DecrementReply 减少回复数
func (r *postRepository) DecrementReply(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&post.Post{}).
		Where("id = ? AND reply_count > 0", id).
		UpdateColumn("reply_count", gorm.Expr("reply_count - ?", 1)).
		Error
}

// Search 搜索帖子
func (r *postRepository) Search(ctx context.Context, keyword string, filter PostFilter) ([]*post.Post, int64, error) {
	var posts []*post.Post
	var total int64

	searchPattern := fmt.Sprintf("%%%s%%", keyword)

	query := r.db.WithContext(ctx).Model(&post.Post{}).Where(
		"title LIKE ? OR content LIKE ? OR company LIKE ? OR location LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// 应用过滤条件
	query = r.applyFilter(query, filter)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序和分页
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "DESC"
	if filter.OrderDir != "" {
		orderDir = filter.OrderDir
	}

	err := query.
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order(fmt.Sprintf("%s %s", orderBy, orderDir)).
		Find(&posts).Error

	return posts, total, err
}

// GetPopular 获取热门帖子（按点赞数排序）
func (r *postRepository) GetPopular(ctx context.Context, offset, limit int) ([]*post.Post, int64, error) {
	var posts []*post.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&post.Post{}).Where("status = ?", post.PostStatusPublished)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按点赞数排序
	err := query.
		Offset(offset).
		Limit(limit).
		Order("like_count DESC, created_at DESC").
		Find(&posts).Error

	return posts, total, err
}

// GetTrending 获取趋势帖子（按浏览数排序）
func (r *postRepository) GetTrending(ctx context.Context, offset, limit int) ([]*post.Post, int64, error) {
	var posts []*post.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&post.Post{}).Where("status = ?", post.PostStatusPublished)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按浏览数排序
	err := query.
		Offset(offset).
		Limit(limit).
		Order("view_count DESC, created_at DESC").
		Find(&posts).Error

	return posts, total, err
}

// applyFilter 应用过滤条件
func (r *postRepository) applyFilter(query *gorm.DB, filter PostFilter) *gorm.DB {
	if filter.PostType != nil {
		query = query.Where("post_type = ?", *filter.PostType)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.AuthorID != nil {
		query = query.Where("author_id = ?", *filter.AuthorID)
	}

	if filter.Location != nil {
		query = query.Where("location = ?", *filter.Location)
	}

	if filter.Company != nil {
		query = query.Where("company = ?", *filter.Company)
	}

	// 标签过滤（JSON数组包含查询）
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
		}
	}

	return query
}
