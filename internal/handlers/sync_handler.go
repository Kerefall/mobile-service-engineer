package handlers

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/Kerefall/mobile-service-engineer/internal/services"
)

type SyncHandler struct {
    syncService *services.SyncService
}

func NewSyncHandler(syncService *services.SyncService) *SyncHandler {
    return &SyncHandler{syncService: syncService}
}

func (h *SyncHandler) SyncOrder(c *gin.Context) {
    orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
        return
    }
    
    var req struct {
        PhotoBefore string `json:"photo_before"`
        PhotoAfter  string `json:"photo_after"`
        Signature   string `json:"signature"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
        return
    }
    
    err = h.syncService.SyncOrder(c.Request.Context(), orderID, req.PhotoBefore, req.PhotoAfter, req.Signature)
    if err != nil {
        logrus.Errorf("Ошибка синхронизации: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка синхронизации"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"success": true, "message": "заказ синхронизирован"})
}