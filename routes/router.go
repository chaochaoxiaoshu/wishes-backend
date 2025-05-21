package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"wishes/controllers"
	"wishes/middleware"
)

type SetupRouterOptions struct {
	UploadController *controllers.UploadController
	AuthController   *controllers.AuthController
	WishController   *controllers.WishController
	RecordController *controllers.RecordController
	UserController   *controllers.UserController
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
				userProtected.GET("/records", options.RecordController.GetWishRecords)
			}
		}

		admin := v1.Group("/admin")
		{
			admin.POST("/register", options.AuthController.AdminRegister)
			admin.POST("/login", options.AuthController.AdminLogin)
		}

		v1.GET("/wishes", options.WishController.GetWishes)

		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			protected.POST("/wishes", options.WishController.CreateWish)
			protected.POST("/wishes/batch", options.WishController.BatchCreateWishes)
			protected.DELETE("/wishes/:id", options.WishController.DeleteWish)
			protected.PUT("/wishes/:id", options.WishController.UpdateWish)
			protected.PUT("/wishes/:id/donor", options.WishController.ClaimWish)

			protected.GET("/records", options.RecordController.GetAllRecords)
			protected.PUT("/records/:id/status", options.RecordController.UpdateRecordStatus)

			protected.GET("/users/admin", options.UserController.GetAdminUsers)
			protected.GET("/users/regular", options.UserController.GetNonAdminUsers)
			protected.PUT("/users/:id/admin", options.UserController.UpdateUserAdmin)

			// 文件上传路由
			protected.POST("/upload/image", options.UploadController.UploadImage)
		}
	}

	return r
}
