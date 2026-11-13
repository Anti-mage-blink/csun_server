package router_handler

import (
	"net/http"
	"strconv"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

// MyApplyQueryHandler 我的发起查询的 Handler
type MyApplyQueryHandler struct {
	service service_repository.MyApplyQueryService
}

// NewMyApplyQueryHandler 创建一个我的发起查询 Handler 实例
func NewMyApplyQueryHandler(service service_repository.MyApplyQueryService) *MyApplyQueryHandler {
	return &MyApplyQueryHandler{service: service}
}

// RegisterMyApplyQueryRoutes 注册我的发起查询相关的路由
func RegisterMyApplyQueryRoutes(r *gin.Engine, h *MyApplyQueryHandler) {
	api := r.Group("/api")
	{
		api.GET("/apply/my_apply_query", h.MyApplyQuery)
	}
}

// MyApplyQuery 处理我的发起查询接口请求
func (h *MyApplyQueryHandler) MyApplyQuery(c *gin.Context) {
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

	data, err := h.service.MyApplyQuery(c.Request.Context(), int32(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询我的发起数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "查询我的发起数据成功",
		"data":    data,
	})
}
