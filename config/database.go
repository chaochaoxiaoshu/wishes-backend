package config

import (
	"fmt"
	"log"
	"time"
	"wishes/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(config *Config, timeZone *time.Location) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(timeZone)
		},
	})

	if err != nil {
		log.Fatalf("无法连接到PostgreSQL数据库: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Wish{},
		&models.WishRecord{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Printf("成功连接到PostgreSQL数据库: %s (时区: %s)\n", config.PostgresName, timeZone.String())
	return db
}
