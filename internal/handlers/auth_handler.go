package handlers

import (
	"net/http"

	"github.com/Kerefall/mobile-service-engineer/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	if err := h.authService.Register(c.Request.Context(), req.Login, req.Password, req.FullName, req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "регистрация успешна"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	engineer, err := h.authService.GetEngineerByLogin(c.Request.Context(), req.Login)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  engineer,
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не авторизован"})
		return
	}

	engineer, err := h.authService.GetEngineerByID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "инженер не найден"})
		return
	}

	c.JSON(http.StatusOK, engineer)
}

func (h *AuthHandler) UpdateFCMToken(c *gin.Context) {
	var req struct {
		FCMToken string `json:"fcm_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не авторизован"})
		return
	}

	userID := userIDInterface.(int64)

	if err := h.authService.UpdateFCMToken(c.Request.Context(), userID, req.FCMToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "FCM токен обновлён"})
}
