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
	"github.com/Kerefall/mobile-service-engineer/internal/services"
	"github.com/Kerefall/mobile-service-engineer/pkg/database"
	"github.com/Kerefall/mobile-service-engineer/pkg/logger"
)

type routeDeps struct {
	cfg                 *config.Config
	db                  *database.PostgresDB
	authHandler         *handlers.AuthHandler
	orderHandler        *handlers.OrderHandler
	partHandler         *handlers.PartHandler
	syncHandler         *handlers.SyncHandler
	integrationHandler  *handlers.IntegrationHandler
	metaHandler         *handlers.MetaHandler
	idemp               gin.HandlerFunc
}

func mountAPI(prefix string, router *gin.Engine, d *routeDeps) {
	api := router.Group(prefix)
	{
		api.POST("/login", d.authHandler.Login)
		api.POST("/register", d.authHandler.Register)

		integration := api.Group("/integration")
		integration.Use(middleware.IntegrationMiddleware(d.cfg))
		{
			integration.POST("/1c/orders", d.integrationHandler.ImportOrdersFromOneC)
		}

		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware(d.cfg.JWTSecret, d.db.Pool))
		{
			authorized.GET("/me", d.authHandler.GetMe)
			authorized.POST("/engineer/fcm-token", d.authHandler.UpdateFCMToken)
			authorized.GET("/meta/client", d.metaHandler.GetClientMeta)

			authorized.GET("/orders", d.orderHandler.GetOrders)
			authorized.GET("/orders/:id", d.orderHandler.GetOrderByID)
			authorized.GET("/orders/:id/navigation", d.orderHandler.GetNavigation)
			authorized.POST("/orders/:id/status", d.orderHandler.UpdateOrderStatus)
			authorized.POST("/orders/:id/photos", d.orderHandler.UploadPhotos)
			authorized.POST("/orders/:id/signature", d.orderHandler.UploadSignature)
			authorized.POST("/orders/:id/voice", d.orderHandler.UploadVoice)
			authorized.GET("/orders/:id/voice", d.orderHandler.ListVoiceNotes)
			authorized.POST("/orders/:id/close", d.orderHandler.CloseOrder)
			authorized.GET("/orders/:id/generate-pdf", d.orderHandler.GeneratePDF)

			authorized.GET("/parts", d.partHandler.GetParts)
			authorized.GET("/parts/lookup", d.partHandler.LookupPart)

			authorized.POST("/orders/:id/parts", d.idemp, d.partHandler.WriteOffParts)
			authorized.POST("/sync", d.idemp, d.syncHandler.SyncOrder)
			authorized.POST("/sync/batch", d.idemp, d.syncHandler.SyncBatch)
			authorized.POST("/orders/:id/sync", d.idemp, d.syncHandler.SyncOrderByPath)

			authorized.POST("/upload", d.orderHandler.UploadFile)
			authorized.POST("/orders", d.orderHandler.CreateOrder)
		}
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

	storageService := services.NewStorageService(cfg)
	authService := services.NewAuthService(db.Pool, cfg.JWTSecret)
	orderService := services.NewOrderService(db.Pool)
	partService := services.NewPartService(db.Pool)
	pdfService := services.NewPDFService(db.Pool)
	onecClient := services.NewOneCClient(cfg)
	fcmService := services.NewFCMService(cfg, authService)
	syncService := services.NewSyncService(db.Pool, storageService, pdfService, partService, onecClient)

	authHandler := handlers.NewAuthHandler(authService)
	orderHandler := handlers.NewOrderHandler(orderService, partService, pdfService, storageService, onecClient, cfg)
	partHandler := handlers.NewPartHandler(partService, orderService)
	syncHandler := handlers.NewSyncHandler(syncService, orderService)
	integrationHandler := handlers.NewIntegrationHandler(orderService, fcmService)
	metaHandler := handlers.NewMetaHandler()

	idemp := middleware.IdempotencyMiddleware(db.Pool)

	deps := &routeDeps{
		cfg:                cfg,
		db:                 db,
		authHandler:        authHandler,
		orderHandler:       orderHandler,
		partHandler:        partHandler,
		syncHandler:        syncHandler,
		integrationHandler: integrationHandler,
		metaHandler:        metaHandler,
		idemp:              idemp,
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key", "X-Integration-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "status": "ok"})
	})

	router.Static("/static", "./uploads")

	mountAPI("/api/v1", router, deps)
	mountAPI("/api", router, deps)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logrus.Infof("Сервер запущен на %s (маршруты /api и /api/v1)", addr)

	if err := router.Run(addr); err != nil {
		logrus.Fatal("Ошибка запуска сервера: ", err)
	}
}