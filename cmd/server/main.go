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

	authService := service.NewAuthService(db.Pool, cfg.JWTSecret)
	orderService := service.NewOrderService(db.Pool)
	partService := service.NewPartService(db.Pool)
	pdfService := service.NewPDFService(db.Pool)
	storageService := service.NewStorageService(cfg)
	syncService := service.NewSyncService(db.Pool, storageService, pdfService)

	authHandler := handlers.NewAuthHandler(authService)
	orderHandler := handlers.NewOrderHandler(orderService, partService, pdfService, storageService)
	partHandler := handlers.NewPartHandler(partService)
	syncHandler := handlers.NewSyncHandler(syncService)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "ok",
		})
	})

	router.Static("/static", "./uploads")

	api := router.Group("/api/v1")
	{
		api.POST("/login", authHandler.Login)
		api.POST("/register", authHandler.Register)

		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware(cfg.JWTSecret, db.Pool))
		{
			authorized.GET("/me", authHandler.GetMe)
			authorized.POST("/engineer/fcm-token", authHandler.UpdateFCMToken)

			authorized.GET("/orders", orderHandler.GetOrders)
			authorized.GET("/orders/:id", orderHandler.GetOrderByID)
			authorized.POST("/orders/:id/status", orderHandler.UpdateOrderStatus)
			authorized.POST("/orders/:id/photos", orderHandler.UploadPhotos)
			authorized.POST("/orders/:id/signature", orderHandler.UploadSignature)
			authorized.POST("/orders/:id/close", orderHandler.CloseOrder)
			authorized.GET("/orders/:id/generate-pdf", orderHandler.GeneratePDF)

			authorized.GET("/parts", partHandler.GetParts)

			idemp := middleware.IdempotencyMiddleware(db.Pool)
			authorized.POST("/orders/:id/parts", idemp, partHandler.WriteOffParts)
			authorized.POST("/sync", idemp, syncHandler.SyncOrder)

			authorized.POST("/upload", orderHandler.UploadFile)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logrus.Infof("Сервер запущен на %s", addr)

	if err := router.Run(addr); err != nil {
		logrus.Fatal("Ошибка запуска сервера: ", err)
	}
}
