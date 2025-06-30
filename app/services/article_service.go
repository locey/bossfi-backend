package services

import (
	"errors"
	"fmt"

	"bossfi-backend/db/database"
	"bossfi-backend/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ArticleService struct{}

func NewArticleService() *ArticleService {
	return &ArticleService{}
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(userID uint, title, content string, images []string) (*models.Article, error) {
	article := &models.Article{
		UserID:  userID,
		Title:   title,
		Content: content,
		Images:  images,
	}

	if err := database.DB.Create(article).Error; err != nil {
		logrus.Errorf("Failed to create article: %v", err)
		return nil, err
	}
	return article, nil
}

// GetArticleByID 根据ID获取文章
func (s *ArticleService) GetArticleByID(id uint) (*models.Article, error) {
	var article models.Article
	err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&article).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("article not found")
		}
		return nil, err
	}
	// 增加浏览量
	database.DB.Model(&article).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	return &article, nil
}

// GetArticles 获取文章列表
func (s *ArticleService) GetArticles(page, pageSize int, sortBy, sortOrder string, userID *uint) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64
	query := database.DB.Model(&models.Article{}).Where("is_deleted = ? ", false)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	if sortBy == "" {
		orderClause = "created_at desc"
	}
	offset := (page - 1) * pageSize
	err := query.Order(orderClause).Offset(offset).Limit(pageSize).Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(id, userID uint, title, content string, images []string) (*models.Article, error) {
	var article models.Article

	err := database.DB.Where("id = ? AND user_id = ? AND is_deleted = ?", id, userID, false).First(&article).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("article not found or not authorized")
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"title":   title,
		"content": content,
		"images":  images,
	}

	if err := database.DB.Model(&article).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &article, nil
}

// DeleteArticle 删除文章（逻辑删除）
func (s *ArticleService) DeleteArticle(id, userID uint) error {
	result := database.DB.Model(&models.Article{}).
		Where("id = ? AND user_id = ? AND is_deleted = ?", id, userID, false).
		Update("is_deleted", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("article not found or not authorized")
	}

	return nil
}

// LikeArticle 点赞文章
func (s *ArticleService) LikeArticle(articleID, userID uint) error {
	// 检查文章是否存在
	var article models.Article
	if err := database.DB.Where("id = ? AND is_deleted = ?", articleID, false).First(&article).Error; err != nil {
		return errors.New("article not found")
	}

	// 检查是否已经点赞
	var existingLike models.ArticleLike
	err := database.DB.Where("article_id = ? AND user_id = ?", articleID, userID).First(&existingLike).Error
	if err == nil {
		return errors.New("already liked")
	}

	// 创建点赞记录
	like := &models.ArticleLike{
		ArticleID: articleID,
		UserID:    userID,
	}

	if err := database.DB.Create(like).Error; err != nil {
		return err
	}

	// 更新文章点赞数
	database.DB.Model(&article).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))

	return nil
}

// UnlikeArticle 取消点赞文章
func (s *ArticleService) UnlikeArticle(articleID, userID uint) error {
	// 删除点赞记录
	result := database.DB.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&models.ArticleLike{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("like not found")
	}

	// 更新文章点赞数
	database.DB.Model(&models.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("like_count", gorm.Expr("like_count - ?", 1))

	return nil
}

// IsArticleLiked 检查用户是否已点赞文章
func (s *ArticleService) IsArticleLiked(articleID, userID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&models.ArticleLike{}).
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Count(&count).Error

	return count > 0, err
}
