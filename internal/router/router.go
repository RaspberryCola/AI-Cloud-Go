package router

import (
	"ai-cloud/internal/controller"
	"ai-cloud/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetUpRouters(r *gin.Engine, uc *controller.UserController, fc *controller.FileController, kc *controller.KBController) {
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
			auth.GET("/path", fc.GetPath)
			auth.GET("/id-path", fc.GetIDPath)
		}
		kb := api.Group("knowledge")
		kb.Use(middleware.JWTAuth())
		{
			// KB
			kb.POST("/create", kc.Create)
			kb.DELETE("/delete", kc.Delete)
			kb.POST("/add", kc.AddExistFile)
			kb.POST("/addNew", kc.AddNewFile)
			kb.GET("/page", kc.PageList)
			kb.GET("/detail", kc.GetKBDetail)
			// Doc
			kb.GET("/docPage", kc.DocPage)
			kb.POST("/docDelete", kc.DeleteDocs)
			// RAG
			kb.POST("/retrieve", kc.Retrieve)
			kb.POST("/chat", kc.Chat)
			kb.POST("/stream", kc.ChatStream)

		}
	}
}
