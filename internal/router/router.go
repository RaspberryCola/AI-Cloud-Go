package router

import (
	"ai-cloud/internal/controller"
	"ai-cloud/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetUpRouters(r *gin.Engine, uc *controller.UserController, fc *controller.FileController, kc *controller.KBController, mc *controller.ModelController, ac *controller.AgentController, cc *controller.ConversationController) {
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
			auth.GET("/idPath", fc.GetIDPath)
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
		model := api.Group("model")
		model.Use(middleware.JWTAuth())
		{
			model.POST("/create", mc.CreateModel)
			model.PUT("/update", mc.UpdateModel)
			model.DELETE("/delete", mc.DeleteModel)
			model.GET("/get", mc.GetModel)
			model.GET("/page", mc.PageModels)
			model.GET("/list", mc.ListModels)
		}
		agent := api.Group("agent")
		agent.Use(middleware.JWTAuth())
		{
			agent.POST("/create", ac.CreateAgent)
			agent.POST("/update", ac.UpdateAgent)
			agent.DELETE("/delete", ac.DeleteAgent)
			agent.GET("/get", ac.GetAgent)
			agent.GET("/page", ac.PageAgents)
			agent.POST("/execute/:id", ac.ExecuteAgent)
			agent.POST("/stream", ac.StreamExecuteAgent)
		}
		conv := api.Group("chat")
		conv.Use(middleware.JWTAuth())
		{
			// 调试模式，不保存历史
			conv.POST("/debug", cc.DebugStreamAgent)
			// 会话相关功能
			conv.POST("/create", cc.CreateConversation)
			conv.POST("/stream", cc.StreamConversation)
			conv.GET("/list", cc.ListConversations)
			conv.GET("/list/agent", cc.ListAgentConversations)
			conv.GET("/history", cc.GetConversationHistory)
			conv.DELETE("/delete", cc.DeleteConversation)
		}
	}
}
