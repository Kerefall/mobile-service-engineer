package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/Kerefall/mobile-service-engineer/internal/config"
	"github.com/Kerefall/mobile-service-engineer/internal/handlers"
	"github.com/Kerefall/mobile-service-engineer/internal/middleware"
	"github.com/Kerefall/mobile-service-engineer/internal/service"
	"github.com/Kerefall/mobile-service-engineer/pkg/database"
	"github.com/Kerefall/mobile-service-engineer/pkg/logger"
)

type routeDeps struct {
	cfg            *config.Config
	db             *database.PostgresDB
	authHandler    *handlers.AuthHandler
	orderHandler   *handlers.OrderHandler
	partHandler    *handlers.PartHandler
	syncHandler    *handlers.SyncHandler
	adminHandler   *handlers.AdminHandler
	chatHandler    *handlers.ChatHandler
	idemp          gin.HandlerFunc
	adminKeyMW     gin.HandlerFunc
}

func mountAPI(prefix string, router *gin.Engine, d *routeDeps) {
	api := router.Group(prefix)
	{
		api.POST("/login", d.authHandler.Login)
		api.POST("/register", d.authHandler.Register)

		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware(d.cfg.JWTSecret, d.db.Pool))
		{
			authorized.GET("/me", d.authHandler.GetMe)
			authorized.POST("/engineer/fcm-token", d.authHandler.UpdateFCMToken)

			authorized.GET("/orders", d.orderHandler.GetOrders)
			authorized.GET("/orders/:id", d.orderHandler.GetOrderByID)
			authorized.POST("/orders/:id/status", d.orderHandler.UpdateOrderStatus)
			authorized.POST("/orders/:id/photos", d.orderHandler.UploadPhotos)
			authorized.POST("/orders/:id/signature", d.orderHandler.UploadSignature)
			authorized.POST("/orders/:id/close", d.orderHandler.CloseOrder)
			authorized.GET("/orders/:id/generate-pdf", d.orderHandler.GeneratePDF)

			authorized.GET("/orders/:id/messages", d.chatHandler.ListMessages)
			authorized.POST("/orders/:id/messages", d.chatHandler.PostMessage)

			authorized.GET("/parts", d.partHandler.GetParts)

			authorized.POST("/orders/:id/parts", d.idemp, d.partHandler.WriteOffParts)
			authorized.POST("/sync", d.idemp, d.syncHandler.SyncOrder)
			authorized.POST("/orders/:id/sync", d.idemp, d.syncHandler.SyncOrderByPath)

			authorized.POST("/upload", d.orderHandler.UploadFile)
		}

		admin := api.Group("/admin")
		admin.Use(d.adminKeyMW)
		admin.POST("/orders", d.adminHandler.CreateOrder)
	}
}

func main() {
	logger.Init()
	cfg := config.Load()

	logrus.Infof("Запуск сервера на порту %s", cfg.ServerPort)

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logrus.Fatal("Ошибка подключения к базе данных: ", err)
	}
	defer db.Close()

	logrus.Info("Подключение к базе данных установлено")

	storageService, err := service.NewStorageService(cfg)
	if err != nil {
		logrus.Fatal("Ошибка инициализации хранилища: ", err)
	}

	authService := service.NewAuthService(db.Pool, cfg.JWTSecret)
	orderService := service.NewOrderService(db.Pool)
	partService := service.NewPartService(db.Pool)
	pdfService := service.NewPDFService(db.Pool)
	syncService := service.NewSyncService(db.Pool, storageService, pdfService)
	notifier := service.NewNotificationService(db.Pool, cfg)
	messageService := service.NewMessageService(db.Pool)

	authHandler := handlers.NewAuthHandler(authService)
	orderHandler := handlers.NewOrderHandler(orderService, partService, pdfService, storageService)
	partHandler := handlers.NewPartHandler(partService)
	syncHandler := handlers.NewSyncHandler(syncService)
	adminHandler := handlers.NewAdminHandler(orderService, notifier)
	chatHandler := handlers.NewChatHandler(messageService, orderService)

	idemp := middleware.IdempotencyMiddleware(db.Pool)
	adminMW := middleware.AdminKeyMiddleware(cfg.AdminSecret)

	deps := &routeDeps{
		cfg:          cfg,
		db:           db,
		authHandler:  authHandler,
		orderHandler: orderHandler,
		partHandler:  partHandler,
		syncHandler:  syncHandler,
		adminHandler: adminHandler,
		chatHandler:  chatHandler,
		idemp:        idemp,
		adminKeyMW:   adminMW,
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key", "X-Admin-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "status": "ok"})
	})

	router.Static("/static", "./uploads")
	router.StaticFile("/admin", "./web/admin.html")

	mountAPI("/api/v1", router, deps)
	mountAPI("/api", router, deps)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logrus.Infof("Сервер запущен на %s (маршруты /api и /api/v1)", addr)

	if err := router.Run(addr); err != nil {
		logrus.Fatal("Ошибка запуска сервера: ", err)
	}
}
