package router

import (
	"ai-cloud/internal/controller"
	"ai-cloud/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetUpRouters(r *gin.Engine, uc *controller.UserController, fc *controller.FileController) {
	api := r.Group("/api")
	{

		publicUser := api.Group("/users")
		{
			publicUser.POST("/register", uc.Register)
			publicUser.POST("/login", uc.Login)
		}

		auth := api.Group("files")
		auth.Use(middleware.JWTAuth())
		{
			auth.POST("/upload", fc.Upload)
			auth.GET("/page", fc.PageList)
			auth.GET("/download", fc.Download)
			auth.DELETE("/delete", fc.Delete)
			auth.POST("/folder", fc.CreateFolder)
			auth.POST("/move", fc.BatchMove)
			auth.GET("/search", fc.Search)
			auth.PUT("/rename", fc.Rename)
		}
	}
}
