package main

import (
	"ai-cloud/config"
	"ai-cloud/internal/controller"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/database"
	"ai-cloud/internal/router"
	"ai-cloud/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func main() {
	config.InitConfig()

	db, _ := database.InitDB()

	userDao := dao.NewUserDao(db)
	userService := service.NewUserService(userDao)
	userController := controller.NewUserController(userService)

	r := gin.Default()
	// 配置路由
	router.SetUpRouters(r, userController)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                                 // 允许所有域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},  // 允许的HTTP方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"}, // 允许的请求头
		ExposeHeaders:    []string{"Content-Length"},                                    // 暴露的响应头
		AllowCredentials: true,                                                          // 允许携带凭证（如Cookie）
		MaxAge:           12 * time.Hour,                                                // 预检请求缓存时间
	}))

	r.Run(":8080")
}
