package router_handler

import (
	"errors"
	"log"
	"net/http"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service_repository.AuthService
}

func NewAuthHandler(authService service_repository.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func RegisterAuthRoutes(r *gin.Engine, authHandler *AuthHandler) {
	api := r.Group("/api")
	{
		api.POST("/login", authHandler.Login)
	}
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

	user, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service_repository.ErrPasswordIncorrect) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "密码错误"})
			return
		}
		if errors.Is(err, service_repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "该用户不存在"})
			return
		}
		if errors.Is(err, service_repository.ErrEmployeeNotFound) {
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
		"data":    user,
	})
}
