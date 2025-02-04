package router

import (
	"ai-cloud/internal/controller"

	"github.com/gin-gonic/gin"
)

func SetUpRouters(r *gin.Engine, uc *controller.UserController) {
	api := r.Group("/api")
	{
		userGroup := api.Group("/users")
		{
			userGroup.POST("/register", uc.Register)
			userGroup.POST("/login", uc.Login)
		}
	}
}
