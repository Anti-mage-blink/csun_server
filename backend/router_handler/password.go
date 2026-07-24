package router_handler

import (
	"errors"
	"net/http"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

type PasswordHandler struct {
	passwordService service_repository.PasswordService
}

func NewPasswordHandler(passwordService service_repository.PasswordService) *PasswordHandler {
	return &PasswordHandler{passwordService: passwordService}
}

func RegisterPasswordRoutes(r *gin.Engine, passwordHandler *PasswordHandler) {
	api := r.Group("/api")
	{
		api.POST("/password/modify", passwordHandler.ModifyPassword)
	}
}

type ModifyPasswordRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *PasswordHandler) ModifyPassword(c *gin.Context) {
	var req ModifyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数不规范"})
		return
	}

	err := h.passwordService.ModifyPassword(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service_repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "该用户不存在"})
			return
		}
		if errors.Is(err, service_repository.ErrPasswordDifferent) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "请设置新密码"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码修改成功",
	})
}
