package routes

import (
	"net/http"

	"checkout-backend/internal/handlers"
	"checkout-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	frontendURL string,
	productHandler *handlers.ProductHandler,
	checkoutHandler *handlers.CheckoutHandler,
) *gin.Engine {
	r := gin.New()

	// Global middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware(frontendURL))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   "ok",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/products", productHandler.GetProducts)
		v1.GET("/products/:id", productHandler.GetProductByID)
		v1.POST("/checkout", checkoutHandler.Checkout)
	}

	return r
}
