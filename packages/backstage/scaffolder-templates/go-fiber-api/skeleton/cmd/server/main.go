package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	{% if values.enableCORS %}"github.com/gofiber/fiber/v2/middleware/cors"{% endif %}
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	{% if values.enableMetrics %}"github.com/gofiber/adaptor/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"{% endif %}
)

func main() {
	// Configuration
	port := getEnv("PORT", "${{ values.port }}")
	logLevel := getEnv("LOG_LEVEL", "info")
	
	log.Printf("Starting ${{ values.name }} service...")
	log.Printf("Log level: %s", logLevel)
	
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "${{ values.name }}",
		ErrorHandler: customErrorHandler,
	})
	
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	
	{% if values.enableCORS %}
	// CORS
	corsOrigins := getEnv("CORS_ORIGINS", "*")
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	{% endif %}
	
	// Routes
	{% if values.enableHealthChecks %}
	app.Get("/health", healthHandler)
	app.Get("/ready", readyHandler)
	{% endif %}
	
	{% if values.enableMetrics %}
	// Prometheus metrics
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	{% endif %}
	
	// API routes
	api := app.Group("/api/v1")
	api.Get("/hello", helloHandler)
	
	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	log.Printf("Server started on port %s", port)
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server stopped gracefully")
}

{% if values.enableHealthChecks %}
func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "healthy",
		"service": "${{ values.name }}",
		"timestamp": time.Now().Unix(),
	})
}

func readyHandler(c *fiber.Ctx) error {
	// Add readiness checks here (database, dependencies, etc.)
	return c.JSON(fiber.Map{
		"status": "ready",
		"service": "${{ values.name }}",
		"timestamp": time.Now().Unix(),
	})
}
{% endif %}

func helloHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Hello from ${{ values.name }}!",
		"version": "1.0.0",
		"timestamp": time.Now().Unix(),
	})
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
		"code": code,
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
