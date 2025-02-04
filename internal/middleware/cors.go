package middleware

import (
	"ai-cloud/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// SetupCORS 封装CORS配置
func SetupCORS() gin.HandlerFunc {
	corsConfig := config.GetConfig().CORS

	maxAge, err := time.ParseDuration(corsConfig.MaxAge)
	if err != nil {
		maxAge = 12 * time.Hour
	}

	return cors.New(cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,     // 允许所有域名
		AllowMethods:     corsConfig.AllowMethods,     // 允许的HTTP方法
		AllowHeaders:     corsConfig.AllowHeaders,     // 允许的请求头
		ExposeHeaders:    corsConfig.ExposeHeaders,    // 暴露的响应头
		AllowCredentials: corsConfig.AllowCredentials, // 允许携带凭证（如Cookie）
		MaxAge:           maxAge,                      // 预检请求缓存时间
	})
}
