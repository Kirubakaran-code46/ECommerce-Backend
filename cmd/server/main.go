package main

import (
	"log"

	"checkout-backend/internal/config"
	"checkout-backend/internal/database"
	"checkout-backend/internal/handlers"
	"checkout-backend/internal/repositories"
	"checkout-backend/internal/routes"
	"checkout-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection pool
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[Server] Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repository layer
	productRepo := repositories.NewProductRepository(db)
	couponRepo := repositories.NewCouponRepository(db)

	// Initialize service layer
	productService := services.NewProductService(productRepo)
	checkoutService := services.NewCheckoutService(productRepo, couponRepo)

	// Initialize HTTP handlers
	productHandler := handlers.NewProductHandler(productService)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)

	// Build Gin router
	router := routes.SetupRouter(cfg.FrontendURL, productHandler, checkoutHandler)

	// Start server
	addr := ":" + cfg.Port
	log.Printf("[Server] Starting HTTP server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("[Server] Server error: %v", err)
	}
}
