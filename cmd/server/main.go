package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/sirupsen/logrus"
    "github.com/Kerefall/mobile-service-engineer/internal/config"
    "github.com/Kerefall/mobile-service-engineer/pkg/database"
    "github.com/Kerefall/mobile-service-engineer/pkg/logger"
    "github.com/Kerefall/mobile-service-engineer/internal/handlers"
    "github.com/Kerefall/mobile-service-engineer/internal/services"
    "github.com/Kerefall/mobile-service-engineer/internal/middleware"
)

func main() {
    // Инициализируем логгер
    logger.Init()
    
    // Загружаем конфиг
    cfg := config.Load()
    
    logrus.Infof("Запуск сервера на порту %s", cfg.ServerPort)
    
    // Подключаемся к базе данных
    db, err := database.NewPostgresDB(cfg)
    if err != nil {
        logrus.Fatal("Ошибка подключения к базе данных: ", err)
    }
    defer db.Close()
    
    logrus.Info("Подключение к базе данных установлено")
    
    // Инициализируем сервисы
    authService := services.NewAuthService(db.Pool, cfg.JWTSecret)
    orderService := services.NewOrderService(db.Pool)
    partService := services.NewPartService(db.Pool)
    pdfService := services.NewPDFService(db.Pool)
    storageService := services.NewStorageService(cfg)
    syncService := services.NewSyncService(db.Pool, storageService, pdfService)
    
    // Инициализируем хендлеры
    authHandler := handlers.NewAuthHandler(authService)
    orderHandler := handlers.NewOrderHandler(orderService, partService, pdfService, storageService)
    partHandler := handlers.NewPartHandler(partService)
    syncHandler := handlers.NewSyncHandler(syncService)
    
    // Создаем роутер
    router := gin.Default()
    
    // Настраиваем CORS
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))
    
    // Health check
    router.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
            "status":  "ok",
        })
    })
    
    // Публичные роуты (без авторизации)
    api := router.Group("/api")
    {
        api.POST("/login", authHandler.Login)
    }
    
    // Защищенные роуты (с JWT авторизацией)
    authorized := api.Group("/")
    authorized.Use(middleware.AuthMiddleware(cfg.JWTSecret, db.Pool))
    {
        // Инженер
        authorized.GET("/me", authHandler.GetMe)
        authorized.POST("/engineer/fcm-token", authHandler.UpdateFCMToken)
        
        // Заказы
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
        authorized.POST("/orders/:id/sync", syncHandler.SyncOrder)
        
        // Загрузка файлов
        authorized.POST("/upload", orderHandler.UploadFile)
    }
    
    // Статические файлы (для доступа к фото, подписям, PDF)
    router.Static("/static", "./uploads")
    
    // Запускаем сервер
    addr := fmt.Sprintf(":%s", cfg.ServerPort)
    logrus.Infof("Сервер запущен на %s", addr)
    logrus.Info("Доступные эндпоинты:")
    logrus.Info("  GET  /ping")
    logrus.Info("  POST /api/login")
    logrus.Info("  GET  /api/me")
    logrus.Info("  GET  /api/orders")
    logrus.Info("  GET  /api/orders/:id")
    logrus.Info("  POST /api/orders/:id/status")
    logrus.Info("  POST /api/orders/:id/photos")
    logrus.Info("  POST /api/orders/:id/signature")
    logrus.Info("  POST /api/orders/:id/close")
    logrus.Info("  GET  /api/orders/:id/generate-pdf")
    logrus.Info("  GET  /api/parts")
    logrus.Info("  POST /api/orders/:id/parts")
    logrus.Info("  POST /api/orders/:id/sync")
    logrus.Info("  POST /api/upload")
    
    if err := router.Run(addr); err != nil {
        logrus.Fatal("Ошибка запуска сервера: ", err)
    }
}