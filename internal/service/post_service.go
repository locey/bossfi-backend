package service

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/domain/post"
	"bossfi-blockchain-backend/internal/domain/user"
	"bossfi-blockchain-backend/internal/repository"
	"bossfi-blockchain-backend/pkg/config"
	"bossfi-blockchain-backend/pkg/logger"
)

// PostService 帖子服务接口
type PostService interface {
	// 帖子管理
	CreatePost(ctx context.Context, userID string, req *CreatePostRequest) (*post.Post, error)
	GetPost(ctx context.Context, postID string) (*PostDetailResponse, error)
	UpdatePost(ctx context.Context, userID string, postID string, req *UpdatePostRequest) (*post.Post, error)
	DeletePost(ctx context.Context, userID string, postID string) error
	PublishPost(ctx context.Context, userID string, postID string) (*post.Post, error)
	ClosePost(ctx context.Context, userID string, postID string) (*post.Post, error)

	// 帖子列表
	ListPosts(ctx context.Context, filter *ListPostsRequest) (*ListPostsResponse, error)
	GetUserPosts(ctx context.Context, userID string, page, pageSize int) (*ListPostsResponse, error)
	SearchPosts(ctx context.Context, req *SearchPostsRequest) (*ListPostsResponse, error)
	GetPopularPosts(ctx context.Context, page, pageSize int) (*ListPostsResponse, error)
	GetTrendingPosts(ctx context.Context, page, pageSize int) (*ListPostsResponse, error)

	// 互动功能
	LikePost(ctx context.Context, userID string, postID string) error
	UnlikePost(ctx context.Context, userID string, postID string) error
	ViewPost(ctx context.Context, postID string) error
}

// 请求和响应结构
type CreatePostRequest struct {
	Title        string        `json:"title" binding:"required,max=255"`
	Content      *string       `json:"content"`
	PostType     post.PostType `json:"post_type" binding:"required"`
	Tags         []string      `json:"tags"`
	Salary       *string       `json:"salary"`
	Location     *string       `json:"location"`
	Company      *string       `json:"company"`
	Requirements *string       `json:"requirements"`
}

type UpdatePostRequest struct {
	Title        *string  `json:"title"`
	Content      *string  `json:"content"`
	Tags         []string `json:"tags"`
	Salary       *string  `json:"salary"`
	Location     *string  `json:"location"`
	Company      *string  `json:"company"`
	Requirements *string  `json:"requirements"`
}

type ListPostsRequest struct {
	PostType *post.PostType   `json:"post_type"`
	Status   *post.PostStatus `json:"status"`
	AuthorID *string          `json:"author_id"`
	Tags     []string         `json:"tags"`
	Location *string          `json:"location"`
	Company  *string          `json:"company"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	OrderBy  string           `json:"order_by"`  // created_at, view_count, like_count, reply_count
	OrderDir string           `json:"order_dir"` // asc, desc
}

type SearchPostsRequest struct {
	Keyword  string           `json:"keyword" binding:"required"`
	PostType *post.PostType   `json:"post_type"`
	Status   *post.PostStatus `json:"status"`
	Tags     []string         `json:"tags"`
	Location *string          `json:"location"`
	Company  *string          `json:"company"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	OrderBy  string           `json:"order_by"`
	OrderDir string           `json:"order_dir"`
}

type PostDetailResponse struct {
	*post.Post
	Author *PostAuthor `json:"author"`
}

type PostAuthor struct {
	ID            string  `json:"id"`
	WalletAddress string  `json:"wallet_address"`
	Username      *string `json:"username"`
	Avatar        *string `json:"avatar"`
}

type ListPostsResponse struct {
	Posts      []*PostDetailResponse `json:"posts"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// postService 帖子服务实现
type postService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
	cfg      *config.Config
	logger   *logger.Logger
}

// NewPostService 创建帖子服务
func NewPostService(
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	logger *logger.Logger,
) PostService {
	return &postService{
		postRepo: postRepo,
		userRepo: userRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

// CreatePost 创建帖子
func (s *postService) CreatePost(ctx context.Context, userID string, req *CreatePostRequest) (*post.Post, error) {
	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查发帖消耗的BOSS币
	postCost := decimal.NewFromFloat(10) // 从配置获取
	if !u.HasSufficientBalance(postCost) {
		return nil, user.ErrInsufficientBalance
	}

	// 创建帖子
	p := &post.Post{
		AuthorID:     userID,
		Title:        req.Title,
		Content:      req.Content,
		PostType:     req.PostType,
		Status:       post.PostStatusDraft,
		Salary:       req.Salary,
		Location:     req.Location,
		Company:      req.Company,
		Requirements: req.Requirements,
		BossCost:     postCost,
	}

	// 添加标签
	if len(req.Tags) > 0 {
		p.Tags = post.Tags(req.Tags)
	}

	// 保存帖子
	if err := s.postRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// 扣除用户BOSS币余额
	if err := u.SubBossBalance(postCost); err != nil {
		s.logger.Error("Failed to deduct boss balance",
			zap.Error(err),
			zap.String("user_id", userID))
		// 这里可以考虑回滚帖子创建
		return nil, fmt.Errorf("failed to deduct balance: %w", err)
	}

	if err := s.userRepo.Update(ctx, u); err != nil {
		s.logger.Error("Failed to update user balance",
			zap.Error(err),
			zap.String("user_id", userID))
		// 这里可以考虑回滚操作
	}

	return p, nil
}

// GetPost 获取帖子详情
func (s *postService) GetPost(ctx context.Context, postID string) (*PostDetailResponse, error) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 获取作者信息
	author, err := s.userRepo.GetByID(ctx, p.AuthorID)
	if err != nil {
		s.logger.Warn("Failed to get post author", zap.Error(err), zap.String("post_id", postID), zap.String("author_id", p.AuthorID))
		// 即使获取作者失败，也返回帖子信息
	}

	response := &PostDetailResponse{
		Post: p,
	}

	if author != nil {
		response.Author = &PostAuthor{
			ID:            author.ID,
			WalletAddress: author.WalletAddress,
			Username:      author.Username,
			Avatar:        author.Avatar,
		}
	}

	return response, nil
}

// UpdatePost 更新帖子
func (s *postService) UpdatePost(ctx context.Context, userID string, postID string, req *UpdatePostRequest) (*post.Post, error) {
	// 获取帖子
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if p.AuthorID != userID {
		return nil, post.ErrUnauthorizedEdit
	}

	// 检查是否可以编辑
	if !p.CanEdit() {
		return nil, post.ErrCannotEdit
	}

	// 更新字段
	if req.Title != nil {
		p.Title = *req.Title
	}
	if req.Content != nil {
		p.Content = req.Content
	}
	if req.Salary != nil {
		p.Salary = req.Salary
	}
	if req.Location != nil {
		p.Location = req.Location
	}
	if req.Company != nil {
		p.Company = req.Company
	}
	if req.Requirements != nil {
		p.Requirements = req.Requirements
	}
	if len(req.Tags) > 0 {
		p.Tags = post.Tags(req.Tags)
	}

	// 保存更新
	if err := s.postRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return p, nil
}

// DeletePost 删除帖子
func (s *postService) DeletePost(ctx context.Context, userID string, postID string) error {
	// 获取帖子
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// 检查权限
	if p.AuthorID != userID {
		return post.ErrUnauthorizedEdit
	}

	// 删除帖子
	return s.postRepo.Delete(ctx, postID)
}

// PublishPost 发布帖子
func (s *postService) PublishPost(ctx context.Context, userID string, postID string) (*post.Post, error) {
	// 获取帖子
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if p.AuthorID != userID {
		return nil, post.ErrUnauthorizedEdit
	}

	// 发布帖子
	if err := p.Publish(); err != nil {
		return nil, err
	}

	// 保存更新
	if err := s.postRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to publish post: %w", err)
	}

	return p, nil
}

// ClosePost 关闭帖子
func (s *postService) ClosePost(ctx context.Context, userID string, postID string) (*post.Post, error) {
	// 获取帖子
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if p.AuthorID != userID {
		return nil, post.ErrUnauthorizedEdit
	}

	// 关闭帖子
	if err := p.Close(); err != nil {
		return nil, err
	}

	// 保存更新
	if err := s.postRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to close post: %w", err)
	}

	return p, nil
}

// ListPosts 获取帖子列表
func (s *postService) ListPosts(ctx context.Context, req *ListPostsRequest) (*ListPostsResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建过滤条件
	filter := repository.PostFilter{
		PostType: req.PostType,
		Status:   req.Status,
		AuthorID: req.AuthorID,
		Tags:     req.Tags,
		Location: req.Location,
		Company:  req.Company,
		Offset:   (req.Page - 1) * req.PageSize,
		Limit:    req.PageSize,
		OrderBy:  req.OrderBy,
		OrderDir: req.OrderDir,
	}

	// 获取帖子列表
	posts, total, err := s.postRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &ListPostsResponse{
		Posts:      make([]*PostDetailResponse, 0, len(posts)),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
	}

	// 获取作者信息
	for _, p := range posts {
		postDetail := &PostDetailResponse{Post: p}

		author, err := s.userRepo.GetByID(ctx, p.AuthorID)
		if err == nil {
			postDetail.Author = &PostAuthor{
				ID:            author.ID,
				WalletAddress: author.WalletAddress,
				Username:      author.Username,
				Avatar:        author.Avatar,
			}
		}

		response.Posts = append(response.Posts, postDetail)
	}

	return response, nil
}

// GetUserPosts 获取用户的帖子
func (s *postService) GetUserPosts(ctx context.Context, userID string, page, pageSize int) (*ListPostsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	posts, total, err := s.postRepo.GetByAuthor(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	author, _ := s.userRepo.GetByID(ctx, userID)

	response := &ListPostsResponse{
		Posts:      make([]*PostDetailResponse, 0, len(posts)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	for _, p := range posts {
		postDetail := &PostDetailResponse{Post: p}

		if author != nil {
			postDetail.Author = &PostAuthor{
				ID:            author.ID,
				WalletAddress: author.WalletAddress,
				Username:      author.Username,
				Avatar:        author.Avatar,
			}
		}

		response.Posts = append(response.Posts, postDetail)
	}

	return response, nil
}

// SearchPosts 搜索帖子
func (s *postService) SearchPosts(ctx context.Context, req *SearchPostsRequest) (*ListPostsResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建过滤条件
	filter := repository.PostFilter{
		PostType: req.PostType,
		Status:   req.Status,
		Tags:     req.Tags,
		Location: req.Location,
		Company:  req.Company,
		Offset:   (req.Page - 1) * req.PageSize,
		Limit:    req.PageSize,
		OrderBy:  req.OrderBy,
		OrderDir: req.OrderDir,
	}

	// 搜索帖子
	posts, total, err := s.postRepo.Search(ctx, req.Keyword, filter)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &ListPostsResponse{
		Posts:      make([]*PostDetailResponse, 0, len(posts)),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
	}

	// 获取作者信息
	for _, p := range posts {
		postDetail := &PostDetailResponse{Post: p}

		author, err := s.userRepo.GetByID(ctx, p.AuthorID)
		if err == nil {
			postDetail.Author = &PostAuthor{
				ID:            author.ID,
				WalletAddress: author.WalletAddress,
				Username:      author.Username,
				Avatar:        author.Avatar,
			}
		}

		response.Posts = append(response.Posts, postDetail)
	}

	return response, nil
}

// GetPopularPosts 获取热门帖子
func (s *postService) GetPopularPosts(ctx context.Context, page, pageSize int) (*ListPostsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	posts, total, err := s.postRepo.GetPopular(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return s.buildPostListResponse(ctx, posts, total, page, pageSize)
}

// GetTrendingPosts 获取趋势帖子
func (s *postService) GetTrendingPosts(ctx context.Context, page, pageSize int) (*ListPostsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	posts, total, err := s.postRepo.GetTrending(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return s.buildPostListResponse(ctx, posts, total, page, pageSize)
}

// LikePost 点赞帖子
func (s *postService) LikePost(ctx context.Context, userID string, postID string) error {
	// TODO: 检查用户是否已经点赞过，避免重复点赞
	// 这里需要实现点赞记录表的逻辑

	return s.postRepo.IncrementLike(ctx, postID)
}

// UnlikePost 取消点赞
func (s *postService) UnlikePost(ctx context.Context, userID string, postID string) error {
	// TODO: 检查用户是否已经点赞过，只有点赞过才能取消
	// 这里需要实现点赞记录表的逻辑

	return s.postRepo.DecrementLike(ctx, postID)
}

// ViewPost 增加浏览次数
func (s *postService) ViewPost(ctx context.Context, postID string) error {
	return s.postRepo.IncrementView(ctx, postID)
}

// buildPostListResponse 构建帖子列表响应
func (s *postService) buildPostListResponse(ctx context.Context, posts []*post.Post, total int64, page, pageSize int) (*ListPostsResponse, error) {
	response := &ListPostsResponse{
		Posts:      make([]*PostDetailResponse, 0, len(posts)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	// 获取作者信息
	for _, p := range posts {
		postDetail := &PostDetailResponse{Post: p}

		author, err := s.userRepo.GetByID(ctx, p.AuthorID)
		if err == nil {
			postDetail.Author = &PostAuthor{
				ID:            author.ID,
				WalletAddress: author.WalletAddress,
				Username:      author.Username,
				Avatar:        author.Avatar,
			}
		}

		response.Posts = append(response.Posts, postDetail)
	}

	return response, nil
}
