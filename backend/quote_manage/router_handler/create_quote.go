package router_handler

import (
	"encoding/json"
	"fmt"
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

// PrepareCreateQuote 获取发起报价单所需的编号及全量基础数据
func (h *CreateQuoteHandler) PrepareCreateQuote(c *gin.Context) {
	data, err := h.createQuoteService.PrepareCreateQuote(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取发起报价单数据成功",
		"data":    data,
	})
}

type SubmitUser struct {
	ID   int32  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// SubmitQuoteInput 组合原生的 quote_manage.Quote，仅重写 AttachmentPathArray 字段以接收前端传来的 JSON 字符串数组 []string
type SubmitQuoteInput struct {
	quote_manage.Quote
	AttachmentPathArray []string `json:"attachment_path_array"`
}

type SubmitQuoteRequest struct {
	Quote      *SubmitQuoteInput          `json:"quote" binding:"required"`
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

	// 提取出内嵌的 Quote 模型实例
	quoteModel := &req.Quote.Quote

	// 将前端发来的 []string 转换为 JSON 字符串存入 Quote.AttachmentPathArray (*string)
	if len(req.Quote.AttachmentPathArray) > 0 {
		bytes, err := json.Marshal(req.Quote.AttachmentPathArray)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("序列化附件相对路径失败: %v", err)})
			return
		}
		str := string(bytes)
		quoteModel.AttachmentPathArray = &str
	} else {
		quoteModel.AttachmentPathArray = nil
	}

	err := h.createQuoteService.SubmitQuote(c.Request.Context(), quoteModel, req.QuoteItems, req.User.ID, req.User.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "提交报价单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "提交报价单成功",
	})
}
