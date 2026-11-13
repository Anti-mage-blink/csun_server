package router_handler

import (
	"net/http"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

// FilingLookHandler 备案查看的 Handler
type FilingLookHandler struct {
	service service_repository.FilingLookService
}

// NewFilingLookHandler 创建一个备案查看 Handler 实例
func NewFilingLookHandler(service service_repository.FilingLookService) *FilingLookHandler {
	return &FilingLookHandler{service: service}
}

// RegisterFilingLookRoutes 注册备案查看相关的路由
func RegisterFilingLookRoutes(r *gin.Engine, h *FilingLookHandler) {
	api := r.Group("/api")
	{
		api.GET("/filing/filing_look", h.FilingLook)
	}
}

// FilingLook 处理备案查看接口请求
func (h *FilingLookHandler) FilingLook(c *gin.Context) {
	data, err := h.service.FilingLook(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询全量备案数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "查询全量备案数据成功",
		"data":    data,
	})
}
