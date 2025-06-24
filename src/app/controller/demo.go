package controller

import (
	"bossfi-backend/src/app/model"
	"bossfi-backend/src/app/service"
	"bossfi-backend/src/core/result"
	"github.com/gin-gonic/gin"
	"strconv"
)

var demoService = service.NewDemoService()

// CreateDemo 创建数据
func CreateDemo(c *gin.Context) {
	var req model.BossfiDemo
	if err := c.ShouldBindJSON(&req); err != nil {
		result.Error(c, result.InvalidParameter)
		return
	}
	if err := demoService.Create(&req); err != nil {
		result.Error(c, result.DBCreateFailed)
		return
	}
	result.OK(c, req)
}

// GetDemoByID 查询数据
func GetDemoByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	demo, err := demoService.GetByID(id)
	if err != nil {
		result.Error(c, result.DBNotExist)
		return
	}
	result.OK(c, demo)
}

// UpdateDemo 更新数据
func UpdateDemo(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.BossfiDemo
	if err := c.ShouldBindJSON(&req); err != nil {
		result.Error(c, result.InvalidParameter)
		return
	}
	req.ID = id
	if err := demoService.Update(&req); err != nil {
		result.Error(c, result.DBUpdateFailed)
		return
	}
	result.OK(c, req)
}

// DeleteDemo 删除数据
func DeleteDemo(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := demoService.Delete(id); err != nil {
		result.Error(c, result.DBDeleteFailed)
		return
	}
	result.OK(c, nil)
}

// ListDemo 查询列表
func ListDemo(c *gin.Context) {
	list, err := demoService.List()
	if err != nil {
		result.Error(c, result.DBQueryFailed)
		return
	}
	result.OK(c, list)
}

// PageDemo 分页查询数据
func PageDemo(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	list, total, err := demoService.Page(page, pageSize)
	if err != nil {
		result.Error(c, result.DBQueryFailed)
		return
	}

	// 返回分页结果
	result.OK(c, gin.H{
		"list":  list,
		"total": total,
	})
}
