package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Configuration holds the server configuration
type Configuration struct {
	Server struct {
		Port        int      `json:"port"`
		Host        string   `json:"host"`
		MetricsPort int      `json:"metrics_port"`
		LogLevel    string   `json:"log_level"`
		Debug       bool     `json:"debug"`
		CORS        CORSConfig `json:"cors"`
	} `json:"server"`
	MCP struct {
		ProtocolVersion string `json:"protocol_version"`
		ServerName      string `json:"server_name"`
		ServerVersion   string `json:"server_version"`
		Capabilities    struct {
			SupportedLanguages []string `json:"supported_languages"`
			SupportsNotebooks  bool     `json:"supports_notebooks"`
			SupportsStreaming  bool     `json:"supports_streaming"`
		} `json:"capabilities"`
	} `json:"mcp"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// MCPResponse represents a standard MCP response
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

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

// loadConfig loads the configuration from environment variables
func loadConfig() Configuration {
	config := Configuration{}
	
	// Set defaults
	config.Server.Port = 9090
	config.Server.Host = "0.0.0.0"
	config.Server.MetricsPort = 9091
	config.Server.LogLevel = "info"
	config.Server.Debug = false
	
	config.Server.CORS.AllowedOrigins = []string{"*"}
	config.Server.CORS.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
	config.Server.CORS.AllowedHeaders = []string{"*"}
	
	config.MCP.ProtocolVersion = "2.0"
	config.MCP.ServerName = getEnv("MCP_SERVER_NAME", "Go MCP Server")
	config.MCP.ServerVersion = getEnv("MCP_SERVER_VERSION", "1.0.0")
	config.MCP.Capabilities.SupportedLanguages = []string{"go", "sql"}
	config.MCP.Capabilities.SupportsNotebooks = true
	config.MCP.Capabilities.SupportsStreaming = true
	
	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// setupRouter sets up the Gin router
func setupRouter(config Configuration) *gin.Engine {
	if !config.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()
	
	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.Server.CORS.AllowedOrigins,
		AllowMethods:     config.Server.CORS.AllowedMethods,
		AllowHeaders:     config.Server.CORS.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Version:   config.MCP.ServerVersion,
			Service:   config.MCP.ServerName,
		})
	})
	
	// Add MCP endpoint
	router.POST("/", handleMCPRequest(config))
	
	// Add SSE endpoint
	router.GET("/sse", handleSSE(config))
	
	return router
}

// handleMCPRequest handles MCP protocol requests
func handleMCPRequest(config Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request map[string]interface{}
		if err := c.BindJSON(&request); err != nil {
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
		
		id, _ := request["id"]
		method, _ := request["method"].(string)
		
		// Handle initialize method
		if method == "initialize" {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				ID:      id,
				Result: map[string]interface{}{
					"capabilities": map[string]interface{}{
						"serverName":    config.MCP.ServerName,
						"serverVersion": config.MCP.ServerVersion,
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
							"supportedLanguages":      config.MCP.Capabilities.SupportedLanguages,
							"supportsNotebooks":       config.MCP.Capabilities.SupportsNotebooks,
							"supportsInlineCompletions": true,
						},
					},
				},
			})
			return
		}
		
		// Forward to centralmind/gateway
		// This is a placeholder - in a real implementation, you would forward the request to the gateway
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result:  map[string]interface{}{"message": "Request forwarded to gateway"},
		})
	}
}

// handleSSE handles Server-Sent Events
func handleSSE(config Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
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
					"serverName":    config.MCP.ServerName,
					"serverVersion": config.MCP.ServerVersion,
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
						"supportedLanguages":      config.MCP.Capabilities.SupportedLanguages,
						"supportsNotebooks":       config.MCP.Capabilities.SupportsNotebooks,
						"supportsInlineCompletions": true,
					},
				},
			},
		}
		
		responseJSON, _ := json.Marshal(response)
		c.SSEvent("", map[string]interface{}{
			"data": string(responseJSON),
		})
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
				c.SSEvent("", map[string]interface{}{
					"type":      "heartbeat",
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"clientId":  clientID,
				})
				c.Writer.Flush()
			case <-done:
				// Client disconnected
				return
			}
		}
	}
}

func main() {
	config := loadConfig()
	
	router := setupRouter(config)
	
	// Start the server
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("Starting MCP server at %s", addr)
	log.Fatal(router.Run(addr))
}
