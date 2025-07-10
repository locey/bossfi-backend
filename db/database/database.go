package database

import (
	"fmt"

	"bossfi-backend/config"
	"bossfi-backend/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		logrus.Errorf("Failed to connect to database: %v", err)
		return err
	}

	logrus.Info("Database connected successfully")

	// 自动迁移模型
	//if err := autoMigrate(); err != nil {
	//	logrus.Errorf("Failed to migrate database: %v", err)
	//	return err
	//}

	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.ArticleCategory{},
		&models.Article{},
		&models.ArticleLike{},
		&models.ArticleComment{},
		&models.ArticleCommentLike{},
	)
}

func GetDB() *gorm.DB {
	return DB
}
