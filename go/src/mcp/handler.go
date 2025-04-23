package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cploutarchou/mcp-servers/go/codeassist"
	"github.com/cploutarchou/mcp-servers/go/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler handles MCP protocol requests
type Handler struct {
	config *config.MCPConfig
	logger *logrus.Logger
	codeAssistHandler *codeassist.Handler
}

// NewHandler creates a new MCP handler
func NewHandler(config *config.MCPConfig, logger *logrus.Logger) *Handler {
	return &Handler{
		config: config,
		logger: logger,
		codeAssistHandler: codeassist.NewHandler(logger),
	}
}

// MCPRequest represents an MCP protocol request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP protocol response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in the MCP protocol
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleMCPRequest handles MCP protocol requests
func (h *Handler) HandleMCPRequest(c *gin.Context) {
	var request MCPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse MCP request")
		c.JSON(http.StatusBadRequest, MCPResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &MCPError{
				Code:    -32700,
				Message: "Parse error",
				Data:    map[string]string{"detail": err.Error()},
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"method": request.Method,
		"id":     request.ID,
	}).Info("Received MCP request")

	// Handle different methods
	switch request.Method {
	case "initialize":
		h.handleInitialize(c, request)
	case "shutdown":
		h.handleShutdown(c, request)
	default:
		h.handleUnknownMethod(c, request)
	}
}

// handleInitialize handles the initialize method
func (h *Handler) handleInitialize(c *gin.Context, request MCPRequest) {
	h.logger.Info("Handling initialize request")

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"capabilities": map[string]interface{}{
				"serverName":    h.config.ServerName,
				"serverVersion": h.config.ServerVersion,
				"tools": []map[string]interface{}{
					{
						"name":        "list_tables",
						"description": "List all available tables",
						"parameters": map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						"name":        "discover_data",
						"description": "Discover data in tables",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"table": map[string]interface{}{
									"type":        "string",
									"description": "Table name to discover",
								},
							},
							"required": []string{"table"},
						},
					},
					{
						"name":        "prepare_query",
						"description": "Prepare a SQL query",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "SQL query to prepare",
								},
							},
							"required": []string{"query"},
						},
					},
					{
						"name":        "query",
						"description": "Execute a SQL query",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "SQL query to execute",
								},
							},
							"required": []string{"query"},
						},
					},
				},
				"capabilities": map[string]interface{}{
					"supportedLanguages":       h.config.Capabilities.SupportedLanguages,
					"supportsNotebooks":        h.config.Capabilities.SupportsNotebooks,
					"supportsInlineCompletions": true,
				},
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleShutdown handles the shutdown method
func (h *Handler) handleShutdown(c *gin.Context, request MCPRequest) {
	h.logger.Info("Handling shutdown request")

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  map[string]interface{}{"success": true},
	}

	c.JSON(http.StatusOK, response)
}

// handleUnknownMethod handles unknown methods
func (h *Handler) handleUnknownMethod(c *gin.Context, request MCPRequest) {
	h.logger.WithField("method", request.Method).Info("Checking if method is a code assistance method")

	// Check if the method is a code assistance method
	if request.Method == "code_completion" || request.Method == "code_analysis" || request.Method == "code_documentation" {
		h.handleCodeAssistRequest(c, request)
		return
	}

	h.logger.WithField("method", request.Method).Warn("Unknown method requested")

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Error: &MCPError{
			Code:    -32601,
			Message: "Method not found",
			Data:    map[string]string{"method": request.Method},
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleCodeAssistRequest handles code assistance requests
func (h *Handler) handleCodeAssistRequest(c *gin.Context, request MCPRequest) {
	h.logger.WithField("method", request.Method).Info("Handling code assistance request")

	result, err := h.codeAssistHandler.HandleMCPRequest(request.Method, request.Params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle code assistance request")
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    map[string]string{"detail": err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	})
}

// HandleSSE handles Server-Sent Events
func (h *Handler) HandleSSE(c *gin.Context) {
	h.logger.Info("Setting up SSE connection")

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Generate a client ID
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Send ready event
	c.SSEvent("", map[string]interface{}{
		"type":     "ready",
		"clientId": clientID,
	})
	c.Writer.Flush()

	// Send capabilities
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"capabilities": map[string]interface{}{
				"serverName":    h.config.ServerName,
				"serverVersion": h.config.ServerVersion,
				"tools": []map[string]interface{}{
					{
						"name":        "list_tables",
						"description": "List all available tables",
						"parameters": map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						"name":        "discover_data",
						"description": "Discover data in tables",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"table": map[string]interface{}{
									"type":        "string",
									"description": "Table name to discover",
								},
							},
							"required": []string{"table"},
						},
					},
					{
						"name":        "prepare_query",
						"description": "Prepare a SQL query",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "SQL query to prepare",
								},
							},
							"required": []string{"query"},
						},
					},
					{
						"name":        "query",
						"description": "Execute a SQL query",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "SQL query to execute",
								},
							},
							"required": []string{"query"},
						},
					},
				},
				"capabilities": map[string]interface{}{
					"supportedLanguages":       h.config.Capabilities.SupportedLanguages,
					"supportsNotebooks":        h.config.Capabilities.SupportsNotebooks,
					"supportsInlineCompletions": true,
				},
			},
		},
	}

	responseJSON, _ := json.Marshal(response)
	c.SSEvent("", string(responseJSON))
	c.Writer.Flush()

	// Keep the connection alive with heartbeats
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Use a channel to detect client disconnection
	done := c.Request.Context().Done()

	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			timestamp := time.Now().UTC().Format(time.RFC3339)
			c.SSEvent("", map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": timestamp,
				"clientId":  clientID,
			})
			c.Writer.Flush()
		case <-done:
			// Client disconnected
			h.logger.WithField("clientId", clientID).Info("SSE client disconnected")
			return
		}
	}
}
