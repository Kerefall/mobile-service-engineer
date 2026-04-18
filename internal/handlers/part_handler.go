package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/gin-gonic/gin"
)

type PartHandler struct {
	partService  *services.PartService
	orderService *services.OrderService
}

func NewPartHandler(partService *services.PartService, orderService *services.OrderService) *PartHandler {
	return &PartHandler{partService: partService, orderService: orderService}
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

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "нет пользователя"})
		return
	}
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заказ не найден"})
		return
	}
	if order.EngineerID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "нет доступа к заказу"})
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

// LookupPart — разбор QR/штрихкода: GET /parts/lookup?q=... (артикул, article:..., id:...).
func (h *PartHandler) LookupPart(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен параметр q"})
		return
	}
	p, err := h.partService.ResolvePartByQRCode(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}
