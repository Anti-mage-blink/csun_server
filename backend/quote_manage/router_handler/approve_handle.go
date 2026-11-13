package router_handler

import (
	"net/http"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

// ApproveHandleHandler 审批处理的 Handler
type ApproveHandleHandler struct {
	service service_repository.ApproveHandleService
}

// NewApproveHandleHandler 创建一个审批处理 Handler 实例
func NewApproveHandleHandler(service service_repository.ApproveHandleService) *ApproveHandleHandler {
	return &ApproveHandleHandler{service: service}
}

// RegisterApproveHandleRoutes 注册审批处理相关的路由
func RegisterApproveHandleRoutes(r *gin.Engine, h *ApproveHandleHandler) {
	api := r.Group("/api")
	{
		api.POST("/approve/approve_handle", h.ApproveHandle)
	}
}

// ApproveHandle 处理审批(同意通过/拒绝退回)接口请求
func (h *ApproveHandleHandler) ApproveHandle(c *gin.Context) {
	var req ApproveHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数格式不正确: " + err.Error()})
		return
	}

	params := &service_repository.ApproveHandleRequestParams{
		Action:    req.Action,
		NodeID:    req.NodeID,
		ProcessID: req.ProcessID,
		Comment:   req.Comment,
	}
	if req.User != nil {
		params.UserID = req.User.ID
		params.UserName = req.User.Name
		params.UserRole = req.User.Role
	}

	err := h.service.ApproveHandle(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "审批操作失败: " + err.Error()})
		return
	}

	var successMessage string
	if req.Action == "approve" {
		successMessage = "审批同意通过处理成功"
	} else {
		successMessage = "审批拒绝退回处理成功"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": successMessage,
	})
}

// ApproveUser 包含操作的用户信息
type ApproveUser struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ApproveHandleRequest 审批处理请求结构体
type ApproveHandleRequest struct {
	Action    string       `json:"action" binding:"required,oneof=approve reject"`
	NodeID    int32        `json:"node_id" binding:"required"`
	ProcessID int32        `json:"process_id" binding:"required"`
	Comment   string       `json:"comment"`
	User      *ApproveUser `json:"user"`
}
