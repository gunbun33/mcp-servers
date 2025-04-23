package codeassist

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler handles code assistance requests
type Handler struct {
	logger              *logrus.Logger
	completionHandler   *CompletionHandler
	analysisHandler     *AnalysisHandler
	documentationHandler *DocumentationHandler
}

// NewHandler creates a new code assistance handler
func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		logger:              logger,
		completionHandler:   NewCompletionHandler(logger),
		analysisHandler:     NewAnalysisHandler(logger),
		documentationHandler: NewDocumentationHandler(logger),
	}
}

// HandleCompletion handles code completion requests
func (h *Handler) HandleCompletion(c *gin.Context) {
	var request CompletionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse completion request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"language": request.Language,
		"line":     request.Line,
		"column":   request.Column,
	}).Info("Received completion request")

	response, err := h.completionHandler.GetCompletions(&request)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get completions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleAnalysis handles code analysis requests
func (h *Handler) HandleAnalysis(c *gin.Context) {
	var request AnalysisRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse analysis request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"language": request.Language,
		"codeSize": len(request.Code),
	}).Info("Received analysis request")

	response, err := h.analysisHandler.AnalyzeCode(&request)
	if err != nil {
		h.logger.WithError(err).Error("Failed to analyze code")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleDocumentation handles documentation requests
func (h *Handler) HandleDocumentation(c *gin.Context) {
	var request DocRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse documentation request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"symbol":   request.Symbol,
		"language": request.Language,
	}).Info("Received documentation request")

	response, err := h.documentationHandler.GetDocumentation(&request)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get documentation")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers code assistance routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	codeAssistGroup := router.Group("/codeassist")
	{
		codeAssistGroup.POST("/completion", h.HandleCompletion)
		codeAssistGroup.POST("/analysis", h.HandleAnalysis)
		codeAssistGroup.POST("/documentation", h.HandleDocumentation)
	}
}

// RegisterMCPTools registers code assistance tools with the MCP protocol
func (h *Handler) RegisterMCPTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "code_completion",
			"description": "Get code completion suggestions",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The code to get completions for",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "The programming language of the code",
					},
					"line": map[string]interface{}{
						"type":        "integer",
						"description": "The line number (0-based) where completions are requested",
					},
					"column": map[string]interface{}{
						"type":        "integer",
						"description": "The column number (0-based) where completions are requested",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Additional context information",
					},
				},
				"required": []string{"code", "language", "line", "column"},
			},
		},
		{
			"name":        "code_analysis",
			"description": "Analyze code for issues and suggestions",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The code to analyze",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "The programming language of the code",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Additional context information",
					},
				},
				"required": []string{"code", "language"},
			},
		},
		{
			"name":        "code_documentation",
			"description": "Get documentation for a code symbol",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "The symbol to get documentation for",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "The programming language of the symbol",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Additional context information",
					},
				},
				"required": []string{"symbol", "language"},
			},
		},
	}
}

// HandleMCPRequest handles MCP protocol requests for code assistance
func (h *Handler) HandleMCPRequest(method string, params json.RawMessage) (interface{}, error) {
	h.logger.WithField("method", method).Info("Handling MCP request for code assistance")

	switch method {
	case "code_completion":
		var request CompletionRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, err
		}
		return h.completionHandler.GetCompletions(&request)
	case "code_analysis":
		var request AnalysisRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, err
		}
		return h.analysisHandler.AnalyzeCode(&request)
	case "code_documentation":
		var request DocRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, err
		}
		return h.documentationHandler.GetDocumentation(&request)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}
