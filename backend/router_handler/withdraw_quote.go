package router_handler

import (
	"net/http"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

// WithdrawQuoteUser 撤回报价单中的用户信息
type WithdrawQuoteUser struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// WithdrawQuoteRequest 撤回报价单请求结构体
type WithdrawQuoteRequest struct {
	ProcessID      int32              `json:"process_id"`
	QuoteProcessID int32              `json:"quote_process_id"` // 兼容 quote_process.id 字段名
	User           *WithdrawQuoteUser `json:"user"`
	UserID         int32              `json:"user_id"`   // 兼容平铺 user_id
	UserName       string             `json:"user_name"` // 兼容平铺 user_name
}

// WithdrawQuoteHandler 撤回报价单 Handler
type WithdrawQuoteHandler struct {
	service service_repository.WithdrawQuoteService
}

// NewWithdrawQuoteHandler 创建撤回报价单 Handler 实例
func NewWithdrawQuoteHandler(service service_repository.WithdrawQuoteService) *WithdrawQuoteHandler {
	return &WithdrawQuoteHandler{service: service}
}

// RegisterWithdrawQuoteRoutes 注册撤回报价单相关的路由
func RegisterWithdrawQuoteRoutes(r *gin.Engine, h *WithdrawQuoteHandler) {
	api := r.Group("/api")
	{
		api.POST("/quote/withdraw_quote", h.WithdrawQuote)
		api.POST("/withdraw_quote", h.WithdrawQuote) // 兼容别名路由
	}
}

// WithdrawQuote 处理撤回报价单接口请求
func (h *WithdrawQuoteHandler) WithdrawQuote(c *gin.Context) {
	var req WithdrawQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数格式不正确: " + err.Error()})
		return
	}

	// 统一解析 ProcessID
	processID := req.ProcessID
	if processID <= 0 {
		processID = req.QuoteProcessID
	}

	// 统一解析 UserID 和 UserName
	var userID int32
	var userName string
	if req.User != nil {
		userID = req.User.ID
		userName = req.User.Name
	}
	if userID <= 0 {
		userID = req.UserID
	}
	if userName == "" {
		userName = req.UserName
	}

	params := &service_repository.WithdrawQuoteRequestParams{
		ProcessID: processID,
		UserID:    userID,
		UserName:  userName,
	}

	err := h.service.WithdrawQuote(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "撤回报价单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "报价单撤回成功",
	})
}
