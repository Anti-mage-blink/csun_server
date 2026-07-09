package router

import (
	"csun_server-backend/handler"

	"github.com/gin-gonic/gin"
)

func RegisterQuoteCreateRoutes(r *gin.Engine, quoteCreateHandler *handler.QuoteCreateHandler) {
	api := r.Group("/api")
	{
		api.GET("/quote/create", quoteCreateHandler.PrepareQuoteCreate)
	}
}
