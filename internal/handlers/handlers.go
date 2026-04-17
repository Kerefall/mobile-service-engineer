package handlers

import (
	"mobile-service-engineer/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	taskSvc *service.TaskService
	authSvc *service.AuthService
}

func NewHandler(ts *service.TaskService, as *service.AuthService) *Handler {
	return &Handler{taskSvc: ts, authSvc: as}
}

func (h *Handler) Register(c *gin.Context) {
	var in struct {
		Name     string `json:"full_name"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	if err := h.authSvc.Register(c.Request.Context(), in.Name, in.Phone, in.Password); err != nil {
		c.JSON(500, gin.H{"error": "registration failed"})
		return
	}
	c.JSON(201, gin.H{"status": "created"})
}

func (h *Handler) Login(c *gin.Context) {
	var in struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&in)
	token, err := h.authSvc.Login(c.Request.Context(), in.Phone, in.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token})
}

func (h *Handler) GetTasks(c *gin.Context) {
	// В продакшене ID берется из Middleware (c.Get("user_id"))
	engID, _ := strconv.Atoi(c.Query("engineer_id"))
	tasks, _ := h.taskSvc.GetTasks(c.Request.Context(), engID)
	c.JSON(200, tasks)
}
