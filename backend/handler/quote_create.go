package handler

import (
	"net/http"

	"csun_server-backend/service"

	"github.com/gin-gonic/gin"
)

type QuoteCreateHandler struct {
	quoteCreateService service.QuoteCreateService
}

func NewQuoteCreateHandler(quoteCreateService service.QuoteCreateService) *QuoteCreateHandler {
	return &QuoteCreateHandler{quoteCreateService: quoteCreateService}
}

// PrepareQuoteCreate 获取新建报价单所需的编号及全量基础数据
func (h *QuoteCreateHandler) PrepareQuoteCreate(c *gin.Context) {
	data, err := h.quoteCreateService.PrepareQuoteCreate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取新建报价单数据成功",
		"data":    data,
	})
}
