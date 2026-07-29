package router_handler

import (
	"net/http"
	"strconv"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

// MyApproveQueryHandler 待审批查询的 Handler
type MyApproveQueryHandler struct {
	service service_repository.MyApproveQueryService
}

// NewMyApproveQueryHandler 创建一个待审批查询 Handler 实例
func NewMyApproveQueryHandler(service service_repository.MyApproveQueryService) *MyApproveQueryHandler {
	return &MyApproveQueryHandler{service: service}
}

// RegisterMyApproveQueryRoutes 注册待审批查询相关的路由
func RegisterMyApproveQueryRoutes(r *gin.Engine, h *MyApproveQueryHandler) {
	api := r.Group("/api")
	{
		api.GET("/approve/my_approve_query", h.MyApproveQuery)
	}
}

// MyApproveQuery 处理待我审批接口请求
func (h *MyApproveQueryHandler) MyApproveQuery(c *gin.Context) {
	// 支持兼容多种入参名称：user_id, userId, id
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		userIDStr = c.Query("userId")
	}
	if userIDStr == "" {
		userIDStr = c.Query("id")
	}

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数缺少用户ID (user_id)"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户ID格式不正确，必须为数字"})
		return
	}

	data, err := h.service.MyApproveQuery(c.Request.Context(), int32(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询待审批数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "查询待审批数据成功",
		"data":    data,
	})
}
