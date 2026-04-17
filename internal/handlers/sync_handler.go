package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kerefall/mobile-service-engineer/internal/dto"
	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SyncHandler struct {
	syncService *services.SyncService
}

func NewSyncHandler(syncService *services.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

func (h *SyncHandler) SyncOrder(c *gin.Context) {
	var req dto.SyncOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	if req.OrderID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен order_id"})
		return
	}

	err := h.syncService.SyncOrder(c.Request.Context(), req.OrderID, req.PhotoBefore, req.PhotoAfter, req.Signature)
	if err != nil {
		logrus.Errorf("Ошибка синхронизации: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка синхронизации"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "заказ синхронизирован"})
}

// SyncOrderByPath — POST /orders/:id/sync (order_id из URL, тело как у /sync без обязательного order_id).
func (h *SyncHandler) SyncOrderByPath(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	var req dto.SyncOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	req.OrderID = id

	err = h.syncService.SyncOrder(c.Request.Context(), req.OrderID, req.PhotoBefore, req.PhotoAfter, req.Signature)
	if err != nil {
		logrus.Errorf("Ошибка синхронизации: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка синхронизации"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "заказ синхронизирован"})
}