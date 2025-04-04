package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"wishes/controllers"
	"wishes/middleware"
)

type SetupRouterOptions struct {
	AuthController *controllers.AuthController
	WishController *controllers.WishController
}

func SetupRouter(options SetupRouterOptions) *gin.Engine {
	r := gin.Default()

	// Logger
	r.Use(middleware.Logger())

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由
	api := r.Group("/api")
	v1 := api.Group("/v1")
	{
		auth := v1.Group("/user")
		{
			auth.POST("/login", options.AuthController.WechatLogin)

			userProtected := auth.Group("/")
			userProtected.Use(middleware.JWTAuth())
			{
				userProtected.GET("/wishes", options.WishController.GetUserDonatedWishes)
			}
		}

		admin := v1.Group("/admin")
		{
			admin.POST("/register", options.AuthController.AdminRegister)
			admin.POST("/login", options.AuthController.AdminLogin)
		}

		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/wishes", options.WishController.GetWishes)
			protected.POST("/wishes", options.WishController.CreateWish)
			protected.DELETE("/wishes/:id", options.WishController.DeleteWish)
			protected.PUT("/wishes/:id", options.WishController.UpdateWish)
			protected.PUT("/wishes/:id/donor", options.WishController.UpdateWishDonor)
		}
	}

	return r
}
