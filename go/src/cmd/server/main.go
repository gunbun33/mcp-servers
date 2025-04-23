package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cploutarchou/mcp-servers/go/codeassist"
	"github.com/cploutarchou/mcp-servers/go/config"
	"github.com/cploutarchou/mcp-servers/go/mcp"
	"github.com/cploutarchou/mcp-servers/go/monitoring"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Server.LogLevel)
	if err == nil {
		logger.SetLevel(level)
	}

	// Set Gin mode
	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CORS.AllowedOrigins,
		AllowMethods:     cfg.Server.CORS.AllowedMethods,
		AllowHeaders:     cfg.Server.CORS.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Create handlers
	mcpHandler := mcp.NewHandler(&cfg.MCP, logger)
	healthHandler := monitoring.NewHealthHandler(&cfg.MCP, logger)
	metricsHandler := monitoring.NewMetricsHandler(logger)

	// Add middleware
	router.Use(metricsHandler.MetricsMiddleware())

	// Add logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration.String(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("Request processed")
	})

	// Register routes
	router.POST("/", mcpHandler.HandleMCPRequest)
	router.GET("/sse", mcpHandler.HandleSSE)
	router.GET("/health", healthHandler.HandleHealthCheck)
	
	// Register code assistance routes
	codeAssistHandler := codeassist.NewHandler(logger)
	codeAssistHandler.RegisterRoutes(router)

	// Create metrics server
	metricsRouter := gin.New()
	metricsRouter.Use(gin.Recovery())
	metricsRouter.GET("/metrics", metricsHandler.HandleMetrics)
	metricsRouter.GET("/health", healthHandler.HandleHealthCheck)

	// Start servers
	mainServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.MetricsPort),
		Handler: metricsRouter,
	}

	// Start servers in goroutines
	go func() {
		logger.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		}).Info("Starting MCP server")

		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start MCP server")
		}
	}()

	go func() {
		logger.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.MetricsPort,
		}).Info("Starting metrics server")

		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start metrics server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown servers
	if err := mainServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("MCP server forced to shutdown")
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Metrics server forced to shutdown")
	}

	logger.Info("Servers exited properly")
}
