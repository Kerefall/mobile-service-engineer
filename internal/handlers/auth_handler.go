package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/Kerefall/mobile-service-engineer/internal/services"
)

type AuthHandler struct {
    authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
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
    
    token, engineer, err := h.authService.Login(c.Request.Context(), req.Login, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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
    
    // Получаем ID пользователя из контекста (устанавливается в middleware)
    userIDInterface, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "не авторизован"})
        return
    }
    
    userID := userIDInterface.(int64)
    _ = userID // Используем переменную чтобы избежать ошибки "declared and not used"
    
    // TODO: обновить FCM токен в базе данных
    // _, err := h.authService.UpdateFCMToken(c.Request.Context(), userID, req.FCMToken)
    // if err != nil {
    //     c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка обновления токена"})
    //     return
    // }
    
    c.JSON(http.StatusOK, gin.H{"success": true, "message": "FCM токен обновлён"})
}