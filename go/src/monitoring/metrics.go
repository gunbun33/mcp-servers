package monitoring

import (
	"fmt"
	
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	logger      *logrus.Logger
	registry    *prometheus.Registry
	requestsTotal *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	activeConnections *prometheus.GaugeVec
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger *logrus.Logger) *MetricsHandler {
	registry := prometheus.NewRegistry()
	
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of MCP requests",
		},
		[]string{"method", "status"},
	)
	
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mcp_request_duration_seconds",
			Help: "Duration of MCP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
	
	activeConnections := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_active_connections",
			Help: "Number of active MCP connections",
		},
		[]string{"type"},
	)
	
	registry.MustRegister(requestsTotal)
	registry.MustRegister(requestDuration)
	registry.MustRegister(activeConnections)
	
	return &MetricsHandler{
		logger:      logger,
		registry:    registry,
		requestsTotal: requestsTotal,
		requestDuration: requestDuration,
		activeConnections: activeConnections,
	}
}

// HandleMetrics handles metrics requests
func (h *MetricsHandler) HandleMetrics(c *gin.Context) {
	h.logger.Info("Handling metrics request")
	handler := promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(c.Writer, c.Request)
}

// IncrementRequestsTotal increments the requests total counter
func (h *MetricsHandler) IncrementRequestsTotal(method, status string) {
	h.requestsTotal.WithLabelValues(method, status).Inc()
}

// ObserveRequestDuration observes the request duration
func (h *MetricsHandler) ObserveRequestDuration(method string, duration float64) {
	h.requestDuration.WithLabelValues(method).Observe(duration)
}

// SetActiveConnections sets the active connections gauge
func (h *MetricsHandler) SetActiveConnections(connectionType string, count float64) {
	h.activeConnections.WithLabelValues(connectionType).Set(count)
}

// IncrementActiveConnections increments the active connections gauge
func (h *MetricsHandler) IncrementActiveConnections(connectionType string) {
	h.activeConnections.WithLabelValues(connectionType).Inc()
}

// DecrementActiveConnections decrements the active connections gauge
func (h *MetricsHandler) DecrementActiveConnections(connectionType string) {
	h.activeConnections.WithLabelValues(connectionType).Dec()
}

// MetricsMiddleware is a middleware that collects metrics for requests
func (h *MetricsHandler) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := prometheus.NewTimer(h.requestDuration.WithLabelValues(c.Request.Method))
		
		c.Next()
		
		start.ObserveDuration()
		h.IncrementRequestsTotal(c.Request.Method, fmt.Sprintf("%d", c.Writer.Status()))
	}
}
