package model

import (
	"bossfi-backend/src/core/db"
	"time"
)

type DemoModel struct {
}

type BossfiDemo struct {
	ID         int64                  `json:"id" gorm:"column:id;primaryKey"`
	Address    string                 `json:"address" gorm:"column:address"`
	Logs       map[string]interface{} `json:"logs" gorm:"column:logs;type:jsonb;serializer:json"`
	Deleted    bool                   `json:"deleted" gorm:"column:deleted"`
	CreateTime time.Time              `json:"create_time" gorm:"column:create_time"`
	ModifyTime time.Time              `json:"modify_time" gorm:"column:modify_time"`
}

func (BossfiDemo) TableName() string {
	return "bossfi_demo"
}

func NewDemoModel() *DemoModel {
	return &DemoModel{}
}

// Create 创建记录
func (m *DemoModel) Create(demo *BossfiDemo) error {
	return db.DB.Create(demo).Error
}

// GetByID 查询单条记录
func (m *DemoModel) GetByID(id int64) (*BossfiDemo, error) {
	var demo BossfiDemo
	err := db.DB.Where("id = ? AND deleted = false", id).First(&demo).Error
	if err != nil {
		return nil, err
	}
	return &demo, nil
}

// Update 更新记录
func (m *DemoModel) Update(demo *BossfiDemo) error {
	return db.DB.Save(demo).Error
}

// Delete 软删除记录
func (m *DemoModel) Delete(id int64) error {
	return db.DB.Model(&BossfiDemo{}).
		Where("id = ?", id).
		Update("deleted", true).Error
}

// List 查询所有未删除记录
func (m *DemoModel) List() ([]*BossfiDemo, error) {
	var list []*BossfiDemo
	err := db.DB.Where("deleted = false").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Page 查询分页数据
func (m *DemoModel) Page(page, pageSize int) ([]*BossfiDemo, int64, error) {
	var list []*BossfiDemo
	var total int64

	res := db.DB.Model(&BossfiDemo{}).Where("deleted = false")

	// 获取总数
	if err := res.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := res.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
