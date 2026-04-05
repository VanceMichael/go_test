package handler

import (
	"net/http"

	"release-manager/internal/config"
	"release-manager/internal/model"
	"release-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
	logger  *config.Logger
}

func NewAuthHandler(authSvc *service.AuthService, logger *config.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	resp, err := h.authSvc.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT无状态，客户端删除token即可
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, user.(*model.User))
}
