package router_handler

import (
	"net/http"

	"csun_server-backend/dao/model/quote_manage"
	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

type CreateQuoteHandler struct {
	createQuoteService service_repository.CreateQuoteService
}

func NewCreateQuoteHandler(createQuoteService service_repository.CreateQuoteService) *CreateQuoteHandler {
	return &CreateQuoteHandler{createQuoteService: createQuoteService}
}

func RegisterCreateQuoteRoutes(r *gin.Engine, createQuoteHandler *CreateQuoteHandler) {
	api := r.Group("/api")
	{
		api.GET("/quote/create", createQuoteHandler.PrepareCreateQuote)
		api.POST("/quote/submit", createQuoteHandler.SubmitQuote)
	}
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

type SubmitUser struct {
	ID   int32  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type SubmitQuoteRequest struct {
	Quote      *quote_manage.Quote        `json:"quote" binding:"required"`
	QuoteItems []*quote_manage.AQuoteItem `json:"quote_items" binding:"required,gt=0"`
	User       *SubmitUser                `json:"user" binding:"required"`
}

// SubmitQuote 提交报价单并写入数据库
func (h *CreateQuoteHandler) SubmitQuote(c *gin.Context) {
	var req SubmitQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数格式不正确: " + err.Error()})
		return
	}

	err := h.createQuoteService.SubmitQuote(c.Request.Context(), req.Quote, req.QuoteItems, req.User.ID, req.User.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "提交报价单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "提交报价单成功",
	})
}
