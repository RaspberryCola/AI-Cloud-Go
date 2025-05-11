package main

import (
	"ai-cloud/config"
	_ "ai-cloud/internal/component/embedding"
	"ai-cloud/internal/controller"
	"ai-cloud/internal/dao"
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

	db, _ := database.InitDB()

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

	agentDao := dao.NewAgentDao(db)
	agentService := service.NewAgentService(agentDao, modelService, kbService, kbDao, modelDao)
	agentController := controller.NewAgentController(agentService)

	r := gin.Default()
	// 配置跨域
	r.Use(middleware.SetupCORS())
	// 配置路由
	router.SetUpRouters(r, userController, fileController, kbController, modelController, agentController)

	r.Run(":8080")
}
