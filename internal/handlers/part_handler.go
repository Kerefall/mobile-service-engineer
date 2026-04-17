package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kerefall/mobile-service-engineer/internal/service"
	"github.com/gin-gonic/gin"
)

type PartHandler struct {
	partService *service.PartService
}

func NewPartHandler(partService *service.PartService) *PartHandler {
	return &PartHandler{partService: partService}
}

func (h *PartHandler) GetParts(c *gin.Context) {
	parts, err := h.partService.GetAllParts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения запчастей"})
		return
	}

	c.JSON(http.StatusOK, parts)
}

func (h *PartHandler) WriteOffParts(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	var req struct {
		Parts []struct {
			PartID   int64 `json:"part_id"`
			Quantity int   `json:"quantity"`
		} `json:"parts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	err = h.partService.WriteOffParts(c.Request.Context(), orderID, req.Parts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Запчасти списаны"})
}
