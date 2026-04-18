package handlers

import (
	"encoding/base64"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
	"strings"

	"github.com/Kerefall/mobile-service-engineer/internal/config"
	"github.com/Kerefall/mobile-service-engineer/internal/models"
	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/Kerefall/mobile-service-engineer/pkg/yandexmaps"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	orderService   *services.OrderService
	partService    *services.PartService
	pdfService     *services.PDFService
	storageService *services.StorageService
	onec           *services.OneCClient
	cfg            *config.Config
}

func NewOrderHandler(orderService *services.OrderService, partService *services.PartService, pdfService *services.PDFService, storageService *services.StorageService, onec *services.OneCClient, cfg *config.Config) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		partService:    partService,
		pdfService:     pdfService,
		storageService: storageService,
		onec:           onec,
		cfg:            cfg,
	}
}

func validWorkflowStatus(s models.OrderStatus) bool {
	switch s {
	case models.StatusNew, models.StatusInProgress, models.StatusCompleted, models.StatusOnTheWay, models.StatusSyncing:
		return true
	default:
		return false
	}
}

// ensureOrderEngineer загружает заказ и проверяет, что он назначен текущему инженеру.
func (h *OrderHandler) ensureOrderEngineer(c *gin.Context, orderID int64) (*models.Order, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "нет пользователя"})
		return nil, false
	}
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заказ не найден"})
		return nil, false
	}
	if order.EngineerID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "нет доступа к заказу"})
		return nil, false
	}
	return order, true
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

	order, ok := h.ensureOrderEngineer(c, id)
	if !ok {
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

	if !validWorkflowStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "недопустимый статус (допустимы: new, in_progress, completed, on_the_way, syncing)"})
		return
	}

	if _, ok := h.ensureOrderEngineer(c, id); !ok {
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

	if _, ok := h.ensureOrderEngineer(c, id); !ok {
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

	capturedAt := time.Now()
	stamped := data
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	if out, newExt, stampErr := services.StampImageWithTime(data, capturedAt); stampErr != nil {
		logrus.Warnf("метка времени на фото: %v", stampErr)
	} else {
		stamped = out
		ext = newExt
	}

	path, err := h.storageService.SaveFile(stamped, "photos", ext)
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

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Фото загружено",
		"path":         path,
		"captured_at":  capturedAt.UTC().Format(time.RFC3339),
	})
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

	if _, ok := h.ensureOrderEngineer(c, id); !ok {
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

const maxVoiceUploadBytes = 10 << 20 // 10 MiB

// UploadVoice загружает голосовое сообщение к заказу (поле multipart: audio).
func (h *OrderHandler) UploadVoice(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}
	if _, ok := h.ensureOrderEngineer(c, id); !ok {
		return
	}

	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл audio не передан"})
		return
	}
	if file.Size > maxVoiceUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл слишком большой (макс. 10 МБ)"})
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
		ext = ".m4a"
	}
	path, err := h.storageService.SaveFile(data, "voice", ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dur *int
	if ds := c.PostForm("duration_sec"); ds != "" {
		if n, e := strconv.Atoi(ds); e == nil && n >= 0 {
			dur = &n
		}
	}

	noteID, err := h.orderService.AddVoiceNote(c.Request.Context(), id, path, dur)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "id": noteID, "path": path})
}

// ListVoiceNotes список голосовых вложений заказа.
func (h *OrderHandler) ListVoiceNotes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}
	if _, ok := h.ensureOrderEngineer(c, id); !ok {
		return
	}

	notes, err := h.orderService.ListVoiceNotes(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notes)
}

func (h *OrderHandler) CloseOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	if _, ok := h.ensureOrderEngineer(c, id); !ok {
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

	full, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось прочитать заказ"})
		return
	}
	partLines, err := h.partService.GetPartsLinesForOrder(c.Request.Context(), id)
	if err != nil {
		logrus.Warnf("запчасти для 1С: %v", err)
		partLines = nil
	}
	voicePaths, _ := h.orderService.VoicePathsForOrder(c.Request.Context(), id)
	doneAt := time.Now()
	payload := services.OrderClosePayload{
		OrderID:            full.ID,
		OneCGuid:           full.OneCGuid,
		Title:              full.Title,
		Address:            full.Address,
		Equipment:          full.Equipment,
		CompletedAt:        doneAt,
		PDFWebPath:         pdfPath,
		PhotoBeforeWebPath: full.PhotoBeforePath,
		PhotoAfterWebPath:  full.PhotoAfterPath,
		SignatureWebPath:   full.SignaturePath,
		VoiceWebPaths:      voicePaths,
		Parts:              partLines,
	}
	pushErr := h.onec.PushOrderClosed(c.Request.Context(), payload)
	if pushErr != nil {
		logrus.Warnf("отправка в 1С: %v", pushErr)
		if h.cfg != nil && h.cfg.OneCFailOnError {
			c.JSON(http.StatusBadGateway, gin.H{"error": "1С: " + pushErr.Error(), "pdf_url": pdfPath})
			return
		}
	}

	if err := h.orderService.UpdateOrderPDFPath(c.Request.Context(), id, pdfPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения пути PDF"})
		return
	}

	if err := h.orderService.CloseOrder(c.Request.Context(), id, pdfPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка закрытия заказа"})
		return
	}

	if pushErr != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Заказ закрыт локально", "pdf_url": pdfPath, "onec_delivered": false, "onec_error": pushErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Заказ закрыт", "pdf_url": pdfPath, "onec_delivered": true})
}

func (h *OrderHandler) GeneratePDF(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	if _, ok := h.ensureOrderEngineer(c, id); !ok {
		return
	}

	pdfPath, err := h.pdfService.GenerateOrderPDF(c.Request.Context(), id, "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка генерации PDF"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "pdf_url": pdfPath})
}

// GetNavigation возвращает ссылку для открытия Яндекс.Карт с маршрутом до объекта заказа.
// Опционально query: from_lat, from_lng — начальная точка маршрута (иначе «с текущего места»).
// Если у заказа нет координат, возвращается ссылка на поиск по адресу.
func (h *OrderHandler) GetNavigation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID заказа"})
		return
	}

	userID, _ := c.Get("user_id")
	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заказ не найден"})
		return
	}
	if order.EngineerID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "нет доступа к заказу"})
		return
	}

	var fromLat, fromLng *float64
	if fs, fsOK := c.GetQuery("from_lat"); fsOK && fs != "" {
		if ls, lsOK := c.GetQuery("from_lng"); lsOK && ls != "" {
			a, e1 := strconv.ParseFloat(fs, 64)
			b, e2 := strconv.ParseFloat(ls, 64)
			if e1 == nil && e2 == nil {
				fromLat = &a
				fromLng = &b
			}
		}
	}

	var navURL string
	mode := "route"
	if order.Latitude == 0 && order.Longitude == 0 {
		navURL = yandexmaps.SearchByAddressURL(order.Address)
		mode = "search_address"
	} else {
		navURL = yandexmaps.RouteToPointURL(order.Latitude, order.Longitude, fromLat, fromLng)
	}

	c.JSON(http.StatusOK, gin.H{
		"yandex_maps_url": navURL,
		"mode":            mode,
		"destination": gin.H{
			"latitude":  order.Latitude,
			"longitude": order.Longitude,
			"address":   order.Address,
		},
	})
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