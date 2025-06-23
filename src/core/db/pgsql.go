package db

import (
	"bossfi-backend/src/core/config"
	"bossfi-backend/src/core/log"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPgsql() *gorm.DB {
	log.Logger.Info("Init Pgsql")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Conf.Pgsql.Host,
		config.Conf.Pgsql.Username,
		config.Conf.Pgsql.Password,
		config.Conf.Pgsql.Database,
		config.Conf.Pgsql.Port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db
	return db
}
