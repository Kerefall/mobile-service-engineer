package handlers

import (
	"encoding/base64"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
	"strings"

	"github.com/Kerefall/mobile-service-engineer/internal/models"
	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	orderService   *services.OrderService
	partService    *services.PartService
	pdfService     *services.PDFService
	storageService *services.StorageService
}

func NewOrderHandler(orderService *services.OrderService, partService *services.PartService, pdfService *services.PDFService, storageService *services.StorageService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		partService:    partService,
		pdfService:     pdfService,
		storageService: storageService,
	}
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.DefaultQuery("status", "active")

	orders, err := h.orderService.GetOrdersByEngineer(c.Request.Context(), userID.(int64), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения заказов"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заказ не найден"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	var req struct {
		Status models.OrderStatus `json:"status" binding:"required"`
		Lat    float64            `json:"lat"`
		Lng    float64            `json:"lng"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	err = h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status, req.Lat, req.Lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка обновления статуса"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Статус обновлён"})
}

func (h *OrderHandler) UploadPhotos(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	photoType := c.PostForm("type")
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "фото не загружено"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка чтения файла"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка чтения файла"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	path, err := h.storageService.SaveFile(data, "photos", ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if photoType == "before" {
		err = h.orderService.UpdateOrderPhotos(c.Request.Context(), id, path, "")
	} else {
		err = h.orderService.UpdateOrderPhotos(c.Request.Context(), id, "", path)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения пути в БД"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Фото загружено", "path": path})
}

func decodeBase64Payload(s string) ([]byte, error) {
	if i := strings.Index(s, ","); i >= 0 {
		s = s[i+1:]
	}
	return base64.StdEncoding.DecodeString(s)
}

func (h *OrderHandler) UploadSignature(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	var req struct {
		Signature string `json:"signature" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	raw, err := decodeBase64Payload(req.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные подписи (base64)"})
		return
	}

	path, err := h.storageService.SaveFile(raw, "signatures", ".png")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.orderService.UpdateOrderSignature(c.Request.Context(), id, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения подписи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Подпись сохранена", "path": path})
}

func (h *OrderHandler) CloseOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	if err := h.orderService.ValidateOrderReadyToClose(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdfPath, err := h.pdfService.GenerateOrderPDF(c.Request.Context(), id, "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка генерации PDF: " + err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderPDFPath(c.Request.Context(), id, pdfPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения пути PDF"})
		return
	}

	err = h.orderService.CloseOrder(c.Request.Context(), id, pdfPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка закрытия заказа"})
		return
	}

	logrus.Infof("[1C] заказ #%d передан в учётную систему (имитация интеграции)", id)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Заказ закрыт", "pdf_url": pdfPath})
}

func (h *OrderHandler) GeneratePDF(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	pdfPath, err := h.pdfService.GenerateOrderPDF(c.Request.Context(), id, "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка генерации PDF"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "pdf_url": pdfPath})
}

func (h *OrderHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не загружен"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка чтения файла"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка чтения файла"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".bin"
	}

	path, err := h.storageService.SaveFile(data, "uploads", ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "path": path})
}

// CreateOrder - создание нового заказа
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req struct {
        Title         string    `json:"title" binding:"required"`
        Description   string    `json:"description"`
        Address       string    `json:"address" binding:"required"`
        ScheduledDate time.Time `json:"scheduled_date" binding:"required"`
        EngineerID    int64     `json:"engineer_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    orderID, err := h.orderService.CreateOrder(c.Request.Context(), req.Title, req.Description, req.Address, req.ScheduledDate, req.EngineerID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "success":  true,
        "order_id": orderID,
        "message":  "заказ создан",
    })
}