package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresName     string

	ServerAddress   string
	JWTSecret       []byte
	WechatAppID     string
	WechatAppSecret string

	// 腾讯云对象存储配置
	COSSecretID   string
	COSSecretKey  string
	COSRegion     string
	COSBucketName string
	COSBaseURL    string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresName := os.Getenv("POSTGRES_NAME")

	serverAddress := os.Getenv("SERVER_ADDRESS")
	jwtSecret := os.Getenv("JWT_SECRET")
	wechatAppId := os.Getenv("WECHAT_APPID")
	wechatAppSecret := os.Getenv("WECHAT_SECRET")

	// 加载腾讯云对象存储配置
	cosSecretID := os.Getenv("COS_SECRET_ID")
	cosSecretKey := os.Getenv("COS_SECRET_KEY")
	cosRegion := os.Getenv("COS_REGION")
	cosBucketName := os.Getenv("COS_BUCKET_NAME")
	cosBaseURL := os.Getenv("COS_BASE_URL")

	return &Config{
		PostgresHost:     postgresHost,
		PostgresPort:     postgresPort,
		PostgresUser:     postgresUser,
		PostgresPassword: postgresPassword,
		PostgresName:     postgresName,

		ServerAddress:   serverAddress,
		JWTSecret:       []byte(jwtSecret),
		WechatAppID:     wechatAppId,
		WechatAppSecret: wechatAppSecret,

		COSSecretID:   cosSecretID,
		COSSecretKey:  cosSecretKey,
		COSRegion:     cosRegion,
		COSBucketName: cosBucketName,
		COSBaseURL:    cosBaseURL,
	}
}
