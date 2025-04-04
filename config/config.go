package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath          string
	ServerAddress   string
	JWTSecret       []byte
	WechatAppID     string
	WechatAppSecret string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	dbPath := os.Getenv("SQLITE_DB_PATH")
	serverAddress := os.Getenv("SERVER_ADDRESS")
	jwtSecret := os.Getenv("JWT_SECRET")
	wechatAppId := os.Getenv("WECHAT_APPID")
	wechatAppSecret := os.Getenv("WECHAT_SECRET")

	return &Config{
		DBPath:          dbPath,
		ServerAddress:   serverAddress,
		JWTSecret:       []byte(jwtSecret),
		WechatAppID:     wechatAppId,
		WechatAppSecret: wechatAppSecret,
	}
}
