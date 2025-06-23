package service

import (
	"bossfi-backend/src/app/model"
)

var dao = model.NewDemoModel()

type DemoService struct{}

func NewDemoService() *DemoService {
	return &DemoService{}
}

// Create 创建记录
func (s *DemoService) Create(demo *model.BossfiDemo) error {
	return dao.Create(demo)
}

// GetByID 查询单条记录
func (s *DemoService) GetByID(id int64) (*model.BossfiDemo, error) {
	return dao.GetByID(id)
}

// Update 更新记录
func (s *DemoService) Update(demo *model.BossfiDemo) error {
	return dao.Update(demo)
}

// Delete 软删除记录
func (s *DemoService) Delete(id int64) error {
	return dao.Delete(id)
}

// List 查询所有未删除记录
func (s *DemoService) List() ([]*model.BossfiDemo, error) {
	return dao.List()
}

// Page 查询分页数据
func (s *DemoService) Page(page, pageSize int) ([]*model.BossfiDemo, int64, error) {
	return dao.Page(page, pageSize)
}
