package router_handler

import (
	"encoding/json"
	"log"
	"net/http"

	"csun_server-backend/hr_sync/service_repository"

	"github.com/gin-gonic/gin"
)

type CallbackRequest struct {
	Key  string          `json:"key" binding:"required"`
	Data json.RawMessage `json:"data"`
}

type CallbackHandler struct {
	service service_repository.CallbackService
}

func NewCallbackHandler(service service_repository.CallbackService) *CallbackHandler {
	return &CallbackHandler{service: service}
}

func RegisterCallbackRoutes(r *gin.Engine, h *CallbackHandler) {
	api := r.Group("/api/hr_sync")
	{
		api.POST("/callback", h.Callback)
	}
}

func (h *CallbackHandler) Callback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[HRSync Callback] 请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, service_repository.CallbackResult{
			ResultCode: service_repository.ResultCodeFail,
			ResultMsg:  "请求参数不规范",
		})
		return
	}

	res, err := h.service.HandleCallback(c.Request.Context(), &service_repository.CallbackParams{
		Key:  req.Key,
		Data: req.Data,
	})
	if err != nil {
		log.Printf("[HRSync Callback] 业务处理失败: key=%s, err=%v", req.Key, err)
		c.JSON(http.StatusInternalServerError, service_repository.CallbackResult{
			ResultCode: service_repository.ResultCodeFail,
			ResultMsg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
