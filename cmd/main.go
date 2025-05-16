package main

import (
	"ai-cloud/config"
	_ "ai-cloud/internal/component/embedding"
	"ai-cloud/internal/controller"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/dao/history"
	"ai-cloud/internal/database"
	"ai-cloud/internal/middleware"
	"ai-cloud/internal/router"
	"ai-cloud/internal/service"
	"context"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()
	ctx := context.Background()

	db, _ := database.GetDB()

	userDao := dao.NewUserDao(db)
	userService := service.NewUserService(userDao)
	userController := controller.NewUserController(userService)
	fileDao := dao.NewFileDao(db)
	fileService := service.NewFileService(fileDao)
	fileController := controller.NewFileController(fileService)

	milvusClient, _ := database.InitMilvus(ctx)
	defer milvusClient.Close()

	modelDao := dao.NewModelDao(db)
	modelService := service.NewModelService(modelDao)
	modelController := controller.NewModelController(modelService)

	kbDao := dao.NewKnowledgeBaseDao(db)
	kbService := service.NewKBService(kbDao, fileService, modelDao)
	kbController := controller.NewKBController(kbService, fileService)

	msgDao := history.NewMsgDao(db)
	convDao := history.NewConvDao(db)
	historyService := service.NewHistoryService(convDao, msgDao)

	agentDao := dao.NewAgentDao(db)
	agentService := service.NewAgentService(agentDao, modelService, kbService, kbDao, modelDao, historyService)
	agentController := controller.NewAgentController(agentService)

	// 创建ConversationService和ConversationController
	conversationService := service.NewConversationService(agentService, historyService)
	conversationController := controller.NewConversationController(conversationService)

	r := gin.Default()
	// 配置跨域
	r.Use(middleware.SetupCORS())
	// 配置路由
	router.SetUpRouters(r, userController, fileController, kbController, modelController, agentController, conversationController)

	r.Run(":8080")
}
