package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"wishes/models"
)

func InitDB(config *Config, timeZone *time.Location) *gorm.DB {
	dir := "./data"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(timeZone)
		},
	})
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	db.AutoMigrate(&models.Wish{}, &models.User{}, &models.Admin{})

	fmt.Printf("成功连接到SQLite数据库: %s (时区: %s)\n", config.DBPath, timeZone.String())
	return db
}
