package config

import (
	"os"
	"time"
)

type Config struct {
	DBPath          string
	ServerAddress   string
	TimeZone        *time.Location
	JWTSecret       []byte
	WechatAppID     string
	WechatAppSecret string
}

func LoadConfig() *Config {
	cst8 := time.FixedZone("CST", 8*3600)
	time.Local = cst8

	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/wishes.db"
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "e8f2c7b5a6d9f0e3d1c4b8a7f2e5d9c6b3a0f1e4d7c8b5a2f9e6d3c0b7a4f1e8"
	}

	wechatAppId := os.Getenv("WECHAT_APPID")
	wechatAppSecret := os.Getenv("WECHAT_SECRET")

	return &Config{
		DBPath:          dbPath,
		ServerAddress:   serverAddress,
		TimeZone:        cst8,
		JWTSecret:       []byte(jwtSecret),
		WechatAppID:     wechatAppId,
		WechatAppSecret: wechatAppSecret,
	}
}
