package services

import (
	"errors"

	"bossfi-backend/db/database"
	"bossfi-backend/models"

	"gorm.io/gorm"
)

type CategoryService struct{}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(name, description, icon, color string, sortOrder int) (*models.ArticleCategory, error) {
	// 检查分类名称是否已存在
	var existingCategory models.ArticleCategory
	if err := database.DB.Where("name = ?", name).First(&existingCategory).Error; err == nil {
		return nil, errors.New("category name already exists")
	}

	category := &models.ArticleCategory{
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		IsActive:    true,
	}

	if err := database.DB.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategoryByID 根据ID获取分类
func (s *CategoryService) GetCategoryByID(id uint) (*models.ArticleCategory, error) {
	var category models.ArticleCategory
	if err := database.DB.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	// 获取文章数量
	var count int64
	database.DB.Model(&models.Article{}).Where("category_id = ? AND is_deleted = ?", id, false).Count(&count)
	category.ArticleCount = count

	return &category, nil
}

// GetCategories 获取分类列表
func (s *CategoryService) GetCategories(page, pageSize int, isActive *bool) ([]models.ArticleCategory, int64, error) {
	var categories []models.ArticleCategory
	var total int64

	query := database.DB.Model(&models.ArticleCategory{})

	// 如果指定了活跃状态，添加筛选条件
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	// 获取每个分类的文章数量
	for i := range categories {
		var count int64
		database.DB.Model(&models.Article{}).
			Where("category_id = ? AND is_deleted = ?", categories[i].ID, false).
			Count(&count)
		categories[i].ArticleCount = count
	}

	return categories, total, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(id uint, name, description, icon, color string, sortOrder int, isActive bool) (*models.ArticleCategory, error) {
	category, err := s.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}

	// 如果修改了名称，检查是否与其他分类重复
	if name != category.Name {
		var existingCategory models.ArticleCategory
		if err := database.DB.Where("name = ? AND id != ?", name, id).First(&existingCategory).Error; err == nil {
			return nil, errors.New("category name already exists")
		}
	}

	// 更新字段
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"icon":        icon,
		"color":       color,
		"sort_order":  sortOrder,
		"is_active":   isActive,
	}

	if err := database.DB.Model(category).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的分类
	return s.GetCategoryByID(id)
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(id uint) error {
	// 检查分类是否存在
	if _, err := s.GetCategoryByID(id); err != nil {
		return err
	}

	// 检查是否有文章使用此分类
	var count int64
	database.DB.Model(&models.Article{}).Where("category_id = ? AND is_deleted = ?", id, false).Count(&count)
	if count > 0 {
		return errors.New("cannot delete category with existing articles")
	}

	// 删除分类
	return database.DB.Delete(&models.ArticleCategory{}, id).Error
}

// GetAllActiveCategories 获取所有活跃分类
func (s *CategoryService) GetAllActiveCategories() ([]models.ArticleCategory, error) {
	var categories []models.ArticleCategory

	if err := database.DB.Where("is_active = ?", true).
		Order("sort_order ASC, created_at DESC").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	// 获取每个分类的文章数量
	for i := range categories {
		var count int64
		database.DB.Model(&models.Article{}).
			Where("category_id = ? AND is_deleted = ?", categories[i].ID, false).
			Count(&count)
		categories[i].ArticleCount = count
	}

	return categories, nil
}
