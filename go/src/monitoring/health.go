package monitoring

import (
	"net/http"
	"time"

	"github.com/cploutarchou/mcp-servers/go/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	config *config.MCPConfig
	logger *logrus.Logger
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(config *config.MCPConfig, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		config: config,
		logger: logger,
	}
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

// HandleHealthCheck handles health check requests
func (h *HealthHandler) HandleHealthCheck(c *gin.Context) {
	h.logger.Info("Handling health check request")

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   h.config.ServerVersion,
		Service:   h.config.ServerName,
	}

	c.JSON(http.StatusOK, response)
}
