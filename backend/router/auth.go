package router

import (
	"csun_server-backend/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	api := r.Group("/api")
	{
		api.POST("/login", authHandler.Login)
	}
}
