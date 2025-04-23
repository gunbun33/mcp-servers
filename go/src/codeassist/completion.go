package codeassist

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// CompletionHandler handles code completion requests
type CompletionHandler struct {
	logger *logrus.Logger
}

// NewCompletionHandler creates a new completion handler
func NewCompletionHandler(logger *logrus.Logger) *CompletionHandler {
	return &CompletionHandler{
		logger: logger,
	}
}

// CompletionRequest represents a code completion request
type CompletionRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Context  string `json:"context,omitempty"`
}

// CompletionItem represents a single completion suggestion
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          string `json:"kind"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText"`
	SortText      string `json:"sortText,omitempty"`
}

// CompletionResponse represents a code completion response
type CompletionResponse struct {
	Items []CompletionItem `json:"items"`
}

// GetCompletions returns code completion suggestions
func (h *CompletionHandler) GetCompletions(request *CompletionRequest) (*CompletionResponse, error) {
	h.logger.WithFields(logrus.Fields{
		"language": request.Language,
		"line":     request.Line,
		"column":   request.Column,
	}).Info("Processing completion request")

	// Extract context from the code
	lines := strings.Split(request.Code, "\n")
	if request.Line >= len(lines) {
		return &CompletionResponse{Items: []CompletionItem{}}, nil
	}

	currentLine := lines[request.Line]
	if request.Column > len(currentLine) {
		request.Column = len(currentLine)
	}

	prefix := currentLine[:request.Column]
	h.logger.WithField("prefix", prefix).Debug("Completion prefix")

	// Generate completions based on language
	var items []CompletionItem
	var err error

	switch strings.ToLower(request.Language) {
	case "go":
		items, err = h.getGoCompletions(request, prefix)
	case "python":
		items, err = h.getPythonCompletions(request, prefix)
	case "sql":
		items, err = h.getSQLCompletions(request, prefix)
	default:
		items, err = h.getGenericCompletions(request, prefix)
	}

	if err != nil {
		return nil, err
	}

	return &CompletionResponse{Items: items}, nil
}

// getGoCompletions returns Go-specific completions
func (h *CompletionHandler) getGoCompletions(request *CompletionRequest, prefix string) ([]CompletionItem, error) {
	// Basic Go keywords and common patterns
	keywords := []string{
		"func", "type", "struct", "interface", "map", "chan", "go", "defer", "if", "else", "for", "range", "switch", "case", "default", "return",
	}

	// Common Go packages
	packages := []string{
		"fmt", "os", "io", "net/http", "encoding/json", "strings", "time", "context", "errors",
	}

	// Common Go methods and functions
	methods := []string{
		"String()", "Error()", "Close()", "Read()", "Write()", "Marshal()", "Unmarshal()", "Print()", "Println()", "Printf()",
	}

	items := []CompletionItem{}

	// Check if we're importing a package
	if strings.Contains(prefix, "import") || strings.Contains(prefix, "\"") {
		for _, pkg := range packages {
			items = append(items, CompletionItem{
				Label:         pkg,
				Kind:          "module",
				Detail:        "Go package",
				Documentation: fmt.Sprintf("Standard library package: %s", pkg),
				InsertText:    pkg,
			})
		}
		return items, nil
	}

	// Check if we're typing a keyword
	for _, keyword := range keywords {
		if strings.HasPrefix(keyword, strings.TrimSpace(prefix)) || prefix == "" {
			items = append(items, CompletionItem{
				Label:         keyword,
				Kind:          "keyword",
				Detail:        "Go keyword",
				Documentation: fmt.Sprintf("Go keyword: %s", keyword),
				InsertText:    keyword,
			})
		}
	}

	// Check if we're calling a method
	if strings.Contains(prefix, ".") {
		for _, method := range methods {
			items = append(items, CompletionItem{
				Label:         method,
				Kind:          "method",
				Detail:        "Go method",
				Documentation: fmt.Sprintf("Common Go method: %s", method),
				InsertText:    method,
			})
		}
	}

	return items, nil
}

// getPythonCompletions returns Python-specific completions
func (h *CompletionHandler) getPythonCompletions(request *CompletionRequest, prefix string) ([]CompletionItem, error) {
	// Basic Python keywords and common patterns
	keywords := []string{
		"def", "class", "if", "else", "elif", "for", "while", "try", "except", "finally", "with", "import", "from", "as", "return", "yield", "lambda",
	}

	// Common Python modules
	modules := []string{
		"os", "sys", "json", "datetime", "math", "random", "re", "collections", "itertools", "functools",
	}

	// Common Python methods and functions
	methods := []string{
		"__init__", "__str__", "__repr__", "append()", "extend()", "pop()", "keys()", "values()", "items()", "get()", "update()",
	}

	items := []CompletionItem{}

	// Check if we're importing a module
	if strings.Contains(prefix, "import") || strings.Contains(prefix, "from") {
		for _, module := range modules {
			items = append(items, CompletionItem{
				Label:         module,
				Kind:          "module",
				Detail:        "Python module",
				Documentation: fmt.Sprintf("Standard library module: %s", module),
				InsertText:    module,
			})
		}
		return items, nil
	}

	// Check if we're typing a keyword
	for _, keyword := range keywords {
		if strings.HasPrefix(keyword, strings.TrimSpace(prefix)) || prefix == "" {
			items = append(items, CompletionItem{
				Label:         keyword,
				Kind:          "keyword",
				Detail:        "Python keyword",
				Documentation: fmt.Sprintf("Python keyword: %s", keyword),
				InsertText:    keyword,
			})
		}
	}

	// Check if we're calling a method
	if strings.Contains(prefix, ".") {
		for _, method := range methods {
			items = append(items, CompletionItem{
				Label:         method,
				Kind:          "method",
				Detail:        "Python method",
				Documentation: fmt.Sprintf("Common Python method: %s", method),
				InsertText:    method,
			})
		}
	}

	return items, nil
}

// getSQLCompletions returns SQL-specific completions
func (h *CompletionHandler) getSQLCompletions(request *CompletionRequest, prefix string) ([]CompletionItem, error) {
	// SQL keywords
	keywords := []string{
		"SELECT", "FROM", "WHERE", "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "GROUP BY", "ORDER BY", "HAVING",
		"INSERT INTO", "VALUES", "UPDATE", "SET", "DELETE FROM", "CREATE TABLE", "ALTER TABLE", "DROP TABLE",
		"AND", "OR", "NOT", "IN", "BETWEEN", "LIKE", "IS NULL", "IS NOT NULL",
	}

	// SQL functions
	functions := []string{
		"COUNT()", "SUM()", "AVG()", "MIN()", "MAX()", "COALESCE()", "NULLIF()", "CAST()", "CONVERT()",
		"UPPER()", "LOWER()", "TRIM()", "SUBSTRING()", "CONCAT()", "LENGTH()", "ROUND()", "NOW()", "CURRENT_DATE()",
	}

	items := []CompletionItem{}

	// Check if we're typing a keyword
	for _, keyword := range keywords {
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(prefix)), strings.Split(keyword, " ")[0]) || prefix == "" {
			items = append(items, CompletionItem{
				Label:         keyword,
				Kind:          "keyword",
				Detail:        "SQL keyword",
				Documentation: fmt.Sprintf("SQL keyword: %s", keyword),
				InsertText:    keyword,
			})
		}
	}

	// Check if we might be using a function
	for _, function := range functions {
		items = append(items, CompletionItem{
			Label:         function,
			Kind:          "function",
			Detail:        "SQL function",
			Documentation: fmt.Sprintf("SQL function: %s", function),
			InsertText:    function,
		})
	}

	return items, nil
}

// getGenericCompletions returns generic completions for unsupported languages
func (h *CompletionHandler) getGenericCompletions(request *CompletionRequest, prefix string) ([]CompletionItem, error) {
	// Generic programming constructs
	constructs := []string{
		"if", "else", "for", "while", "function", "class", "return", "var", "let", "const",
	}

	items := []CompletionItem{}

	for _, construct := range constructs {
		if strings.HasPrefix(construct, strings.TrimSpace(prefix)) || prefix == "" {
			items = append(items, CompletionItem{
				Label:         construct,
				Kind:          "keyword",
				Detail:        "Programming construct",
				Documentation: fmt.Sprintf("Common programming construct: %s", construct),
				InsertText:    construct,
			})
		}
	}

	return items, nil
}

// ToJSON converts the completion response to JSON
func (r *CompletionResponse) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
