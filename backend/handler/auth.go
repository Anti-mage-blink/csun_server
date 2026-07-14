package handler

import (
	"errors"
	"log"
	"net/http"

	"csun_server-backend/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数不规范"})
		return
	}

	employee, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrPasswordIncorrect) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "密码错误"})
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "该用户不存在"})
			return
		}
		if errors.Is(err, service.ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "关联的员工不存在"})
			return
		}
		// 记录真正的错误原因，便于在容器日志中排查
		log.Printf("[Login] 内部错误: username=%s, err=%v", req.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data":    employee,
	})
}
