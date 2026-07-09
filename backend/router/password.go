package router

import (
	"csun_server-backend/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPasswordRoutes(r *gin.Engine, passwordHandler *handler.PasswordHandler) {
	api := r.Group("/api")
	{
		api.POST("/password/modify", passwordHandler.ModifyPassword)
	}
}
