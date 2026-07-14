package router

import (
	"csun_server-backend/handler"

	"github.com/gin-gonic/gin"
)

func RegisterCreateQuoteRoutes(r *gin.Engine, createQuoteHandler *handler.CreateQuoteHandler) {
	api := r.Group("/api")
	{
		api.GET("/quote/create", createQuoteHandler.PrepareCreateQuote)
	}
}
