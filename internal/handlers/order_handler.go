package handlers

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/Kerefall/mobile-service-engineer/internal/models"
    "github.com/Kerefall/mobile-service-engineer/internal/services"
)

type OrderHandler struct {
    orderService  *services.OrderService
    partService   *services.PartService
    pdfService    *services.PDFService
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
    
    // Читаем файл
    src, err := file.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка чтения файла"})
        return
    }
    defer src.Close()
    
    // Сохраняем
    var path string
    if photoType == "before" {
        filename := h.storageService.GenerateUniqueFilename("order_before", ".jpg")
        // TODO: сохранить файл
        path = "/static/photos/" + filename
        h.orderService.UpdateOrderPhotos(c.Request.Context(), id, path, "")
    } else {
        filename := h.storageService.GenerateUniqueFilename("order_after", ".jpg")
        path = "/static/photos/" + filename
        h.orderService.UpdateOrderPhotos(c.Request.Context(), id, "", path)
    }
    
    c.JSON(http.StatusOK, gin.H{"success": true, "message": "Фото загружено", "path": path})
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
    
    filename := h.storageService.GenerateUniqueFilename("signature", ".png")
    path := "/static/signatures/" + filename
    
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
    
    // Генерируем PDF
    pdfPath, err := h.pdfService.GenerateOrderPDF(c.Request.Context(), id, "", "", "")
    if err != nil {
        pdfPath = ""
    }
    
    err = h.orderService.CloseOrder(c.Request.Context(), id, pdfPath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка закрытия заказа"})
        return
    }
    
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
    
    filename := h.storageService.GenerateUniqueFilename("upload", "")
    // TODO: сохранить файл
    
    c.JSON(http.StatusOK, gin.H{"success": true, "path": "/static/uploads/" + filename})
}