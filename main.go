package main

import (
	"time"
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
	cst8 := time.FixedZone("CST", 8*3600)
	time.Local = cst8

	cfg := config.LoadConfig()
	middleware.InitJWTSecret(cfg.JWTSecret)
	db := config.InitDB(cfg, cst8)

	// 初始化服务
	wechatService := services.NewWechatService(db, cfg.WechatAppID, cfg.WechatAppSecret, cfg.JWTSecret)
	wishService := services.NewWishService(db)
	recordService := services.NewRecordService(db)
	userService := services.NewUserService(db)

	// 初始化控制器
	authController := controllers.NewAuthController(db, wechatService)
	wishController := controllers.NewWishController(wishService, recordService, userService)
	recordController := controllers.NewRecordController(recordService)

	// 设置路由
	r := routes.SetupRouter(routes.SetupRouterOptions{
		AuthController:   authController,
		WishController:   wishController,
		RecordController: recordController,
	})

	r.Run(cfg.ServerAddress)
}
