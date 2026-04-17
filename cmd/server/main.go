package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	// Убедись, что пути соответствуют твоему go.mod
	"github.com/Kerefall/mobile-service-engineer/internal/config"
	"github.com/Kerefall/mobile-service-engineer/internal/handlers"
	"github.com/Kerefall/mobile-service-engineer/internal/middleware"
	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/Kerefall/mobile-service-engineer/pkg/database"
	"github.com/Kerefall/mobile-service-engineer/pkg/logger"
)

func main() {
	// 1. Инициализация инфраструктуры
	logger.Init()
	cfg := config.Load()

	logrus.Infof("Запуск сервера на порту %s", cfg.ServerPort)

	// 2. Подключение к базе данных
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logrus.Fatal("Ошибка подключения к базе данных: ", err)
	}
	defer db.Close()

	logrus.Info("Подключение к базе данных установлено")

	// 3. Инициализация сервисов (Business Logic)
	// Используем db.Pool, если твоя библиотека БД его предоставляет
	authService := services.NewAuthService(db.Pool, cfg.JWTSecret)
	orderService := services.NewOrderService(db.Pool)
	partService := services.NewPartService(db.Pool)
	pdfService := services.NewPDFService(db.Pool)
	storageService := services.NewStorageService(cfg)
	syncService := services.NewSyncService(db.Pool, storageService, pdfService)

	// 4. Инициализация хендлеров (Controllers)
	authHandler := handlers.NewAuthHandler(authService)
	orderHandler := handlers.NewOrderHandler(orderService, partService, pdfService, storageService)
	partHandler := handlers.NewPartHandler(partService)
	syncHandler := handlers.NewSyncHandler(syncService)

	// 5. Настройка роутера
	router := gin.Default()

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "ok",
		})
	})

	// Статические файлы (фото, подписи, PDF)
	router.Static("/static", "./uploads")

	// --- РОУТИНГ ---

	api := router.Group("/api/v1")
	{
		// Публичные роуты
		api.POST("/login", authHandler.Login)
		api.POST("/register", authHandler.Register) // Если метод реализован

		// Защищенные роуты
		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware(cfg.JWTSecret, db.Pool))
		{
			// Инженер
			authorized.GET("/me", authHandler.GetMe)
			authorized.POST("/engineer/fcm-token", authHandler.UpdateFCMToken)

			// Заказы (Tasks)
			authorized.GET("/orders", orderHandler.GetOrders)
			authorized.GET("/orders/:id", orderHandler.GetOrderByID)
			authorized.POST("/orders/:id/status", orderHandler.UpdateOrderStatus)
			authorized.POST("/orders/:id/photos", orderHandler.UploadPhotos)
			authorized.POST("/orders/:id/signature", orderHandler.UploadSignature)
			authorized.POST("/orders/:id/close", orderHandler.CloseOrder)
			authorized.GET("/orders/:id/generate-pdf", orderHandler.GeneratePDF)

			// Запчасти
			authorized.GET("/parts", partHandler.GetParts)
			authorized.POST("/orders/:id/parts", partHandler.WriteOffParts)

			// Синхронизация (офлайн-режим)
			authorized.POST("/sync", syncHandler.SyncOrder)

			// Загрузка файлов
			authorized.POST("/upload", orderHandler.UploadFile)
		}
	}

	// 6. Запуск сервера
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logrus.Infof("Сервер запущен на %s", addr)

	if err := router.Run(addr); err != nil {
		logrus.Fatal("Ошибка запуска сервера: ", err)
	}
}
