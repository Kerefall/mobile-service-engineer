package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/gin-gonic/gin"
)

type PartHandler struct {
	partService *services.PartService
}

func NewPartHandler(partService *services.PartService) *PartHandler {
	return &PartHandler{partService: partService}
}

func (h *PartHandler) GetParts(c *gin.Context) {
	parts, err := h.partService.GetParts(c.Request.Context())
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
			PartID   int64 `json:"part_id" binding:"required"`
			Quantity int   `json:"quantity" binding:"required,min=1"`
		} `json:"parts" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	parts := make([]struct{ PartID int64; Quantity int }, len(req.Parts))
	for i, p := range req.Parts {
		parts[i].PartID = p.PartID
		parts[i].Quantity = p.Quantity
	}

	err = h.partService.WriteOffParts(c.Request.Context(), orderID, parts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "запчасти списаны"})
}