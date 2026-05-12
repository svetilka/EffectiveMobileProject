package main

import (
	//"net/http"
	"fmt"

	"github.com/svetilka/EffectiveMobileProject/internal/config"
	"github.com/svetilka/EffectiveMobileProject/internal/database"
	"github.com/svetilka/EffectiveMobileProject/internal/handlers"
	"github.com/svetilka/EffectiveMobileProject/internal/repository"

	_ "github.com/svetilka/EffectiveMobileProject/docs"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	config.SetupLogger(cfg.LogLevel)

	log.WithFields(log.Fields{
		"db_host": cfg.DBHost,
		"db_name": cfg.DBName,
		"port":    cfg.ServerPort,
	}).Info("Starting subscription service")

	// Connect to database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.WithError(err).Warn("Migration warning (may be non-critical)")
	}

	// Setup repository and handler
	repo := repository.NewSubscriptionRepository(db.DB)
	handler := handlers.NewSubscriptionHandler(repo)

	// Setup router
	router := setupRouter(handler)

	// Start server
	serverAddr := cfg.GetServerAddress()
	log.WithField("address", serverAddr).Info("Server starting")

	if err := router.Run(serverAddr); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

func setupRouter(handler *handlers.SubscriptionHandler) *gin.Engine {
	router := gin.Default()

	// Add logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	router.Use(func(c *gin.Context) {
		log.WithFields(log.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		}).Info("Incoming request")

		c.Next()

		log.WithFields(log.Fields{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
		}).Info("Request completed")
	})

	// API routes
	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", handler.CreateSubscription)
			subscriptions.GET("", handler.ListSubscriptions)
			subscriptions.GET("/total-cost", handler.CalculateTotalCost)
			subscriptions.GET("/:id", handler.GetSubscription)
			subscriptions.PUT("/:id", handler.UpdateSubscription)
			subscriptions.DELETE("/:id", handler.DeleteSubscription)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	setupSwagger(router)

	return router
}

func setupSwagger(router *gin.Engine) {
	// Simple swagger UI redirect
	router.GET("/swagger/*any", func(c *gin.Context) {
		c.String(200, "Swagger documentation available at /docs/swagger.yaml")
	})
}
