package handler

import (
	"net/http"

	"csun_server-backend/service"

	"github.com/gin-gonic/gin"
)

type CreateQuoteHandler struct {
	createQuoteService service.CreateQuoteService
}

func NewCreateQuoteHandler(createQuoteService service.CreateQuoteService) *CreateQuoteHandler {
	return &CreateQuoteHandler{createQuoteService: createQuoteService}
}

// PrepareCreateQuote 获取新建报价单所需的编号及全量基础数据
func (h *CreateQuoteHandler) PrepareCreateQuote(c *gin.Context) {
	data, err := h.createQuoteService.PrepareCreateQuote(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取新建报价单数据成功",
		"data":    data,
	})
}
