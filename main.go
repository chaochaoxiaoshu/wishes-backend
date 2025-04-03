package main

import (
	"wishes/config"
	"wishes/controllers"
	_ "wishes/docs"
	"wishes/middleware"
	"wishes/routes"
	"wishes/services"
)

// @title           心愿墙 API
// @version         1.0
// @description     心愿墙公益项目API
// @host      localhost:8080
func main() {
	cfg := config.LoadConfig()
	middleware.InitJWTSecret(cfg.JWTSecret)

	db := config.InitDB(cfg)

	// 初始化服务
	wechatService := services.NewWechatService(db, cfg.WechatAppID, cfg.WechatAppSecret, cfg.JWTSecret)

	// 初始化控制器
	authController := controllers.NewAuthController(db, wechatService)
	wishController := controllers.NewWishController(db)

	// 设置路由
	r := routes.SetupRouter(routes.SetupRouterOptions{
		AuthController: authController,
		WishController: wishController,
	})

	r.Run(cfg.ServerAddress)
}
