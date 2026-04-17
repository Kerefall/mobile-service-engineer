package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/sirupsen/logrus"
    "github.com/Kerefall/mobile-service-engineer/internal/config"
    "github.com/Kerefall/mobile-service-engineer/pkg/database"
    "github.com/Kerefall/mobile-service-engineer/pkg/logger"
)

func main() {
    // Инициализируем логгер
    logger.Init()
    
    // Загружаем конфиг
    cfg := config.Load()
    
    logrus.Infof("Starting server on port %s", cfg.ServerPort)
    
    // Подключаемся к БД
    db, err := database.NewPostgresDB(cfg)
    if err != nil {
        logrus.Fatal("Failed to connect to database: ", err)
    }
    defer db.Close()
    
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
    
    // Запускаем сервер
    addr := fmt.Sprintf(":%s", cfg.ServerPort)
    logrus.Infof("Server listening on %s", addr)
    
    if err := router.Run(addr); err != nil {
        logrus.Fatal("Failed to start server: ", err)
    }
}