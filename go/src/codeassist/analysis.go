package codeassist

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// AnalysisHandler handles code analysis requests
type AnalysisHandler struct {
	logger *logrus.Logger
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler(logger *logrus.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		logger: logger,
	}
}

// AnalysisRequest represents a code analysis request
type AnalysisRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Context  string `json:"context,omitempty"`
}

// AnalysisDiagnostic represents a diagnostic issue found in code
type AnalysisDiagnostic struct {
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error", "warning", "info", "hint"
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	EndLine  int    `json:"endLine,omitempty"`
	EndCol   int    `json:"endColumn,omitempty"`
	Code     string `json:"code,omitempty"`
}

// AnalysisSuggestion represents a suggestion for improving code
type AnalysisSuggestion struct {
	Message     string `json:"message"`
	Replacement string `json:"replacement,omitempty"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	EndLine     int    `json:"endLine,omitempty"`
	EndCol      int    `json:"endColumn,omitempty"`
}

// AnalysisResponse represents a code analysis response
type AnalysisResponse struct {
	Diagnostics []AnalysisDiagnostic `json:"diagnostics"`
	Suggestions []AnalysisSuggestion `json:"suggestions"`
	Summary     string               `json:"summary"`
}

// AnalyzeCode analyzes code and returns diagnostics and suggestions
func (h *AnalysisHandler) AnalyzeCode(request *AnalysisRequest) (*AnalysisResponse, error) {
	h.logger.WithFields(logrus.Fields{
		"language": request.Language,
		"codeSize": len(request.Code),
	}).Info("Processing analysis request")

	// Process based on language
	var diagnostics []AnalysisDiagnostic
	var suggestions []AnalysisSuggestion
	var err error

	switch strings.ToLower(request.Language) {
	case "go":
		diagnostics, suggestions, err = h.analyzeGoCode(request.Code)
	case "python":
		diagnostics, suggestions, err = h.analyzePythonCode(request.Code)
	case "sql":
		diagnostics, suggestions, err = h.analyzeSQLCode(request.Code)
	default:
		diagnostics, suggestions, err = h.analyzeGenericCode(request.Code)
	}

	if err != nil {
		return nil, err
	}

	// Generate a summary
	summary := h.generateSummary(diagnostics, suggestions)

	return &AnalysisResponse{
		Diagnostics: diagnostics,
		Suggestions: suggestions,
		Summary:     summary,
	}, nil
}

// analyzeGoCode analyzes Go code
func (h *AnalysisHandler) analyzeGoCode(code string) ([]AnalysisDiagnostic, []AnalysisSuggestion, error) {
	diagnostics := []AnalysisDiagnostic{}
	suggestions := []AnalysisSuggestion{}

	lines := strings.Split(code, "\n")

	// Check for common Go issues
	for i, line := range lines {
		lineNum := i + 1

		// Check for unused imports
		if strings.Contains(line, "import") && strings.Contains(line, "\"") {
			importName := strings.TrimSpace(strings.Split(strings.Split(line, "\"")[1], "\"")[0])
			if !strings.Contains(code, importName+".") && importName != "_" && !strings.Contains(importName, "/") {
				diagnostics = append(diagnostics, AnalysisDiagnostic{
					Message:  fmt.Sprintf("Unused import: %s", importName),
					Severity: "warning",
					Line:     lineNum,
					Column:   strings.Index(line, "\"") + 1,
				})
			}
		}

		// Check for error handling
		if strings.Contains(line, "err :=") || strings.Contains(line, ", err :=") || strings.Contains(line, ", err =") {
			nextLine := ""
			if i+1 < len(lines) {
				nextLine = lines[i+1]
			}
			if !strings.Contains(nextLine, "if err != nil") && !strings.Contains(nextLine, "if err == nil") {
				suggestions = append(suggestions, AnalysisSuggestion{
					Message:     "Missing error handling",
					Replacement: line + "\n\tif err != nil {\n\t\treturn err\n\t}",
					Line:        lineNum,
					Column:      0,
				})
			}
		}

		// Check for naked returns
		if strings.TrimSpace(line) == "return" {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "Naked return should be avoided for readability",
				Severity: "info",
				Line:     lineNum,
				Column:   0,
			})
		}

		// Check for context handling in functions
		if strings.Contains(line, "func") && strings.Contains(line, "(") && !strings.Contains(line, "context.Context") {
			if strings.Contains(code, "http.") || strings.Contains(code, "net/http") {
				suggestions = append(suggestions, AnalysisSuggestion{
					Message: "Consider adding context.Context as the first parameter for better cancellation and timeout handling",
					Line:    lineNum,
					Column:  0,
				})
			}
		}
	}

	return diagnostics, suggestions, nil
}

// analyzePythonCode analyzes Python code
func (h *AnalysisHandler) analyzePythonCode(code string) ([]AnalysisDiagnostic, []AnalysisSuggestion, error) {
	diagnostics := []AnalysisDiagnostic{}
	suggestions := []AnalysisSuggestion{}

	lines := strings.Split(code, "\n")

	// Check for common Python issues
	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Check for print statements in Python 3 code
		if strings.HasPrefix(trimmedLine, "print ") && !strings.Contains(trimmedLine, "(") {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "Python 3 requires parentheses for print function",
				Severity: "error",
				Line:     lineNum,
				Column:   strings.Index(line, "print") + 1,
			})
			suggestions = append(suggestions, AnalysisSuggestion{
				Message:     "Use print() function syntax",
				Replacement: strings.Replace(line, "print ", "print(", 1) + ")",
				Line:        lineNum,
				Column:      0,
			})
		}

		// Check for bare excepts
		if trimmedLine == "except:" {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "Bare 'except:' should be avoided as it catches all exceptions including KeyboardInterrupt",
				Severity: "warning",
				Line:     lineNum,
				Column:   0,
			})
			suggestions = append(suggestions, AnalysisSuggestion{
				Message:     "Specify exception type",
				Replacement: "except Exception:",
				Line:        lineNum,
				Column:      0,
			})
		}

		// Check for mutable default arguments
		if strings.Contains(line, "def ") && (strings.Contains(line, "=[]") || strings.Contains(line, "= []") || strings.Contains(line, "={}") || strings.Contains(line, "= {}")) {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "Mutable default argument can cause unexpected behavior",
				Severity: "warning",
				Line:     lineNum,
				Column:   0,
			})
			suggestions = append(suggestions, AnalysisSuggestion{
				Message: "Use None as default and initialize in function body",
				Line:    lineNum,
				Column:  0,
			})
		}

		// Check for unused imports
		if strings.HasPrefix(trimmedLine, "import ") || strings.HasPrefix(trimmedLine, "from ") {
			importName := ""
			if strings.HasPrefix(trimmedLine, "import ") {
				importName = strings.TrimSpace(strings.Split(trimmedLine, " ")[1])
				if strings.Contains(importName, " as ") {
					importName = strings.TrimSpace(strings.Split(importName, " as ")[1])
				}
			} else if strings.HasPrefix(trimmedLine, "from ") {
				if strings.Contains(trimmedLine, " import ") {
					importParts := strings.Split(trimmedLine, " import ")
					if len(importParts) > 1 {
						importName = strings.TrimSpace(importParts[1])
						if importName == "*" {
							importName = ""  // Skip wildcard imports
						}
					}
				}
			}
			
			if importName != "" && !strings.Contains(code, importName+".") && !strings.Contains(code, " "+importName+" ") && !strings.Contains(code, "("+importName+")") {
				diagnostics = append(diagnostics, AnalysisDiagnostic{
					Message:  fmt.Sprintf("Unused import: %s", importName),
					Severity: "warning",
					Line:     lineNum,
					Column:   0,
				})
			}
		}
	}

	return diagnostics, suggestions, nil
}

// analyzeSQLCode analyzes SQL code
func (h *AnalysisHandler) analyzeSQLCode(code string) ([]AnalysisDiagnostic, []AnalysisSuggestion, error) {
	diagnostics := []AnalysisDiagnostic{}
	suggestions := []AnalysisSuggestion{}

	// Check for common SQL issues
	
	// Check for SELECT *
	if strings.Contains(strings.ToUpper(code), "SELECT *") {
		lineNum := 0
		for i, line := range strings.Split(code, "\n") {
			if strings.Contains(strings.ToUpper(line), "SELECT *") {
				lineNum = i + 1
				break
			}
		}
		
		diagnostics = append(diagnostics, AnalysisDiagnostic{
			Message:  "Using SELECT * can impact performance and may return unnecessary columns",
			Severity: "warning",
			Line:     lineNum,
			Column:   strings.Index(strings.ToUpper(strings.Split(code, "\n")[lineNum-1]), "SELECT *") + 1,
		})
		suggestions = append(suggestions, AnalysisSuggestion{
			Message: "Specify only the columns you need",
			Line:    lineNum,
			Column:  0,
		})
	}
	
	// Check for missing WHERE clause in UPDATE or DELETE
	updateOrDeleteRegex := regexp.MustCompile(`(?i)(UPDATE|DELETE FROM)\s+\w+\s+(SET\s+.*\s+)?(ORDER BY|LIMIT|$)`)
	if updateOrDeleteRegex.MatchString(code) {
		lineNum := 0
		for i, line := range strings.Split(code, "\n") {
			if updateOrDeleteRegex.MatchString(line) {
				lineNum = i + 1
				break
			}
		}
		
		diagnostics = append(diagnostics, AnalysisDiagnostic{
			Message:  "UPDATE or DELETE without WHERE clause will affect all rows",
			Severity: "error",
			Line:     lineNum,
			Column:   0,
		})
		suggestions = append(suggestions, AnalysisSuggestion{
			Message: "Add a WHERE clause to limit the scope of the operation",
			Line:    lineNum,
			Column:  0,
		})
	}
	
	// Check for potential SQL injection
	if strings.Contains(code, "+") && (strings.Contains(strings.ToUpper(code), "WHERE") || strings.Contains(strings.ToUpper(code), "VALUES")) {
		lineNum := 0
		for i, line := range strings.Split(code, "\n") {
			if strings.Contains(line, "+") && (strings.Contains(strings.ToUpper(line), "WHERE") || strings.Contains(strings.ToUpper(line), "VALUES")) {
				lineNum = i + 1
				break
			}
		}
		
		diagnostics = append(diagnostics, AnalysisDiagnostic{
			Message:  "String concatenation in SQL queries can lead to SQL injection vulnerabilities",
			Severity: "error",
			Line:     lineNum,
			Column:   0,
		})
		suggestions = append(suggestions, AnalysisSuggestion{
			Message: "Use parameterized queries or prepared statements instead of string concatenation",
			Line:    lineNum,
			Column:  0,
		})
	}

	return diagnostics, suggestions, nil
}

// analyzeGenericCode analyzes code in unsupported languages
func (h *AnalysisHandler) analyzeGenericCode(code string) ([]AnalysisDiagnostic, []AnalysisSuggestion, error) {
	diagnostics := []AnalysisDiagnostic{}
	suggestions := []AnalysisSuggestion{}

	lines := strings.Split(code, "\n")

	// Check for common issues in any language
	for i, line := range lines {
		lineNum := i + 1
		
		// Check for very long lines
		if len(line) > 100 {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "Line exceeds 100 characters which may affect readability",
				Severity: "info",
				Line:     lineNum,
				Column:   100,
			})
		}
		
		// Check for TODO comments
		if strings.Contains(strings.ToUpper(line), "TODO") {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "TODO comment found",
				Severity: "info",
				Line:     lineNum,
				Column:   strings.Index(strings.ToUpper(line), "TODO"),
			})
		}
		
		// Check for hardcoded credentials
		credentialRegex := regexp.MustCompile(`(?i)(password|passwd|pwd|secret|key|token|api_key|apikey)\s*[=:]\s*['"][^'"]*['"]`)
		if credentialRegex.MatchString(line) {
			diagnostics = append(diagnostics, AnalysisDiagnostic{
				Message:  "Potential hardcoded credential detected",
				Severity: "warning",
				Line:     lineNum,
				Column:   0,
			})
			suggestions = append(suggestions, AnalysisSuggestion{
				Message: "Use environment variables or a secure configuration system for sensitive information",
				Line:    lineNum,
				Column:  0,
			})
		}
	}

	return diagnostics, suggestions, nil
}

// generateSummary generates a summary of the analysis
func (h *AnalysisHandler) generateSummary(diagnostics []AnalysisDiagnostic, suggestions []AnalysisSuggestion) string {
	errorCount := 0
	warningCount := 0
	infoCount := 0
	
	for _, diag := range diagnostics {
		switch diag.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}
	
	summary := fmt.Sprintf("Found %d errors, %d warnings, and %d informational issues. ", errorCount, warningCount, infoCount)
	
	if len(suggestions) > 0 {
		summary += fmt.Sprintf("Provided %d suggestions for improvement.", len(suggestions))
	}
	
	return summary
}

// ToJSON converts the analysis response to JSON
func (r *AnalysisResponse) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
