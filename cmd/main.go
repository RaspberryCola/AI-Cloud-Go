package main

import (
	"ai-cloud/config"
	"ai-cloud/internal/controller"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/database"
	"ai-cloud/internal/middleware"
	"ai-cloud/internal/router"
	"ai-cloud/internal/service"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.InitConfig()

	db, _ := database.InitDB()

	userDao := dao.NewUserDao(db)
	userService := service.NewUserService(userDao)
	userController := controller.NewUserController(userService)
	fileDao := dao.NewFileDao(db)
	fileService := service.NewFileService(fileDao)
	fileController := controller.NewFileController(fileService)

	milvus, _ := database.InitMilvus(context.Background())
	milvusDao := dao.NewMilvusDao(milvus)

	kbDao := dao.NewKnowledgeBaseDao(db)
	kbService := service.NewKBService(kbDao, milvusDao, fileService)
	kbController := controller.NewKBController(kbService, fileService)

	r := gin.Default()
	// 配置跨域
	r.Use(middleware.SetupCORS())
	// 配置路由
	router.SetUpRouters(r, userController, fileController, kbController)

	r.Run(":8080")
}
