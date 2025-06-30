package services

import (
	"errors"

	"bossfi-backend/db/database"
	"bossfi-backend/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ArticleCommentService struct{}

func NewArticleCommentService() *ArticleCommentService {
	return &ArticleCommentService{}
}

// CreateComment 创建评论
func (s *ArticleCommentService) CreateComment(userID, articleID uint, parentID *uint, content string) (*models.ArticleComment, error) {
	// 检查文章是否存在
	var article models.Article
	if err := database.DB.Where("id = ? AND is_deleted = ?", articleID, false).First(&article).Error; err != nil {
		return nil, errors.New("article not found")
	}

	// 如果是回复评论，检查父评论是否存在
	if parentID != nil {
		var parentComment models.ArticleComment
		if err := database.DB.Where("id = ? AND article_id = ? AND is_deleted = ?", parentID, articleID, false).First(&parentComment).Error; err != nil {
			return nil, errors.New("parent comment not found")
		}
	}

	comment := &models.ArticleComment{
		UserID:    userID,
		ArticleID: articleID,
		ParentID:  parentID,
		Content:   content,
	}

	if err := database.DB.Create(comment).Error; err != nil {
		logrus.Errorf("Failed to create comment: %v", err)
		return nil, err
	}

	// 更新文章评论数
	database.DB.Model(&article).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1))

	return comment, nil
}

// GetComments 获取文章评论列表
func (s *ArticleCommentService) GetComments(articleID uint, page, pageSize int, parentID *uint) ([]models.ArticleComment, int64, error) {
	var comments []models.ArticleComment
	var total int64

	query := database.DB.Model(&models.ArticleComment{}).
		Where("article_id = ? AND is_deleted = ?", articleID, false)

	// 如果指定了父评论ID，只查询该评论的回复
	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	} else {
		// 否则只查询顶级评论
		query = query.Where("parent_id IS NULL")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Preload("Replies.User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error

	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// UpdateComment 更新评论
func (s *ArticleCommentService) UpdateComment(id, userID uint, content string) (*models.ArticleComment, error) {
	var comment models.ArticleComment

	err := database.DB.Where("id = ? AND user_id = ? AND is_deleted = ?", id, userID, false).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found or not authorized")
		}
		return nil, err
	}

	if err := database.DB.Model(&comment).Update("content", content).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

// DeleteComment 删除评论（逻辑删除）
func (s *ArticleCommentService) DeleteComment(id, userID uint) error {
	var comment models.ArticleComment

	err := database.DB.Where("id = ? AND user_id = ? AND is_deleted = ?", id, userID, false).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("comment not found or not authorized")
		}
		return err
	}

	// 逻辑删除评论
	if err := database.DB.Model(&comment).Update("is_deleted", true).Error; err != nil {
		return err
	}

	// 更新文章评论数
	database.DB.Model(&models.Article{}).
		Where("id = ?", comment.ArticleID).
		UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1))

	return nil
}

// LikeComment 点赞评论
func (s *ArticleCommentService) LikeComment(commentID, userID uint) error {
	// 检查评论是否存在
	var comment models.ArticleComment
	if err := database.DB.Where("id = ? AND is_deleted = ?", commentID, false).First(&comment).Error; err != nil {
		return errors.New("comment not found")
	}

	// 检查是否已经点赞
	var existingLike models.ArticleCommentLike
	err := database.DB.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&existingLike).Error
	if err == nil {
		return errors.New("already liked")
	}

	// 创建点赞记录
	like := &models.ArticleCommentLike{
		CommentID: commentID,
		UserID:    userID,
	}

	if err := database.DB.Create(like).Error; err != nil {
		return err
	}

	// 更新评论点赞数
	database.DB.Model(&comment).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))

	return nil
}

// UnlikeComment 取消点赞评论
func (s *ArticleCommentService) UnlikeComment(commentID, userID uint) error {
	// 删除点赞记录
	result := database.DB.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&models.ArticleCommentLike{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("like not found")
	}

	// 更新评论点赞数
	database.DB.Model(&models.ArticleComment{}).
		Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("like_count - ?", 1))

	return nil
}

// IsCommentLiked 检查用户是否已点赞评论
func (s *ArticleCommentService) IsCommentLiked(commentID, userID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&models.ArticleCommentLike{}).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Count(&count).Error

	return count > 0, err
}
