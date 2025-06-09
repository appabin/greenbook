package config

import (
	"log"
	"time"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() {
	dsn := AppConfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(AppConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(AppConfig.Database.MaxIdOpenCons)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 执行自动迁移
	err = db.AutoMigrate(
		&models.User{},
		&models.UserFollow{},
		&models.Article{},
		&models.Tag{},
		&models.Like{},
		&models.Comment{},
		&models.Picture{},
		&models.ArticlePicture{},
		&models.Favorite{},
		&models.CommentLike{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	global.Db = db

}
