package main

import (
	"ai-cloud/config"
	"ai-cloud/internal/controller"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/database"
	"ai-cloud/internal/middleware"
	"ai-cloud/internal/router"
	"ai-cloud/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms/openai"
	"log"
)

func main() {
	config.InitConfig()

	db, _ := database.InitDB()

	userDao := dao.NewUserDao(db)
	userService := service.NewUserService(userDao)
	userController := controller.NewUserController(userService)
	fileDao := dao.NewFileDao(db)
	fileService := service.NewFileService(fileDao)
	fileController := controller.NewFileController(fileService)

	opts := []openai.Option{
		openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		openai.WithEmbeddingModel("text-embedding-v3"),
		openai.WithToken("sk-98077dd2f6d74722ba818a4d52e6dee9"),
	}
	embedder, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	kbDao := dao.NewKnowledgeBaseDao(db)
	kbService := service.NewKBService(kbDao, embedder, fileService)
	kbController := controller.NewKBCotroller(kbService, fileService)

	r := gin.Default()
	// 配置跨域
	r.Use(middleware.SetupCORS())
	// 配置路由
	router.SetUpRouters(r, userController, fileController, kbController)

	r.Run(":8080")
}
