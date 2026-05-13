package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/svetilka/EffectiveMobileProject/docs" // swagger docs
)

func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", handler.CreateSubscription)
			subscriptions.GET("", handler.ListSubscriptions)
			subscriptions.GET("/stats", handler.CalculateTotalCost)
			subscriptions.GET("/:id", handler.GetSubscription)
			subscriptions.PUT("/:id", handler.UpdateSubscription)
			subscriptions.DELETE("/:id", handler.DeleteSubscription)
		}
	}

	// Swagger endpoint
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
