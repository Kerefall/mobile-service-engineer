package handlers

import (
	"net/http"
	"strconv"

	"mobile-service-engineer/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler объединяет все зависимости для API
type Handler struct {
	taskSvc *service.TaskService
	authSvc *service.AuthService
	fileSvc *service.FileService
	pdfSvc  *service.PDFService
}

// NewHandler создает новый экземпляр со всеми сервисами
func NewHandler(ts *service.TaskService, as *service.AuthService, fs *service.FileService, ps *service.PDFService) *Handler {
	return &Handler{
		taskSvc: ts,
		authSvc: as,
		fileSvc: fs,
		pdfSvc:  ps,
	}
}

// Register - Эндпоинт регистрации
func (h *Handler) Register(c *gin.Context) {
	var input struct {
		FullName string `json:"full_name" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Все поля обязательны"})
		return
	}

	err := h.authSvc.Register(c.Request.Context(), input.FullName, input.Phone, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при регистрации"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Регистрация успешна"})
}

// Login - Эндпоинт входа
func (h *Handler) Login(c *gin.Context) {
	var input struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	token, err := h.authSvc.Login(c.Request.Context(), input.Phone, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetTasks — получение списка заказ-нарядов
func (h *Handler) GetTasks(c *gin.Context) {
	// Получаем ID инженера из query-параметра (например, /tasks?engineer_id=1)
	engIDStr := c.Query("engineer_id")
	engID, err := strconv.Atoi(engIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID инженера должен быть числом"})
		return
	}

	tasks, err := h.taskSvc.GetTasks(c.Request.Context(), engID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении списка задач"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// UploadPhoto — загрузка фотографий "До" или "После"
func (h *Handler) UploadPhoto(c *gin.Context) {
	// Лимит на файл (например, 10 МБ) задается в настройках Gin в main
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл фото не найден в запросе"})
		return
	}
	defer file.Close()

	taskID := c.PostForm("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID задачи (task_id)"})
		return
	}

	// Сохраняем файл на диск через сервис
	path, err := h.fileSvc.SaveUpload(file, header, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить файл"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"path":   path,
	})
}

// GenerateAct — генерация итогового PDF акта
func (h *Handler) GenerateAct(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID задачи"})
		return
	}

	// Здесь вызываем сервис генерации PDF.
	// Мы передаем ID, сервис должен сам подтянуть данные задачи и фото из БД.
	path, err := h.pdfSvc.GenerateTaskAct(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при генерации PDF: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Акт успешно сформирован",
		"url":     path,
	})
}
