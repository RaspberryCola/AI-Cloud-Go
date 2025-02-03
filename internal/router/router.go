package router

import (
	"ai-cloud/internal/controller"
	"github.com/gin-gonic/gin"
)

func SetUserRouter(uc *controller.UserController) *gin.Engine {
	r := gin.Default()
	// 用户相关路由
	api := r.Group("/api")
	{
		userGroup := api.Group("/users")
		{
			userGroup.POST("/register", uc.Register)
			userGroup.POST("/login", uc.Login)
		}
	}
	return r
}
