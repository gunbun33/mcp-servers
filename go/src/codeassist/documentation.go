package codeassist

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// DocumentationHandler handles documentation requests
type DocumentationHandler struct {
	logger *logrus.Logger
}

// NewDocumentationHandler creates a new documentation handler
func NewDocumentationHandler(logger *logrus.Logger) *DocumentationHandler {
	return &DocumentationHandler{
		logger: logger,
	}
}

// DocRequest represents a documentation request
type DocRequest struct {
	Symbol   string `json:"symbol"`
	Language string `json:"language"`
	Context  string `json:"context,omitempty"`
}

// DocResponse represents a documentation response
type DocResponse struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Syntax      string `json:"syntax,omitempty"`
	Example     string `json:"example,omitempty"`
	URL         string `json:"url,omitempty"`
}

// GetDocumentation returns documentation for a symbol
func (h *DocumentationHandler) GetDocumentation(request *DocRequest) (*DocResponse, error) {
	h.logger.WithFields(logrus.Fields{
		"symbol":   request.Symbol,
		"language": request.Language,
	}).Info("Processing documentation request")

	// Get documentation based on language
	switch strings.ToLower(request.Language) {
	case "go":
		return h.getGoDocumentation(request.Symbol)
	case "python":
		return h.getPythonDocumentation(request.Symbol)
	case "sql":
		return h.getSQLDocumentation(request.Symbol)
	default:
		return h.getGenericDocumentation(request.Symbol, request.Language)
	}
}

// getGoDocumentation returns Go-specific documentation
func (h *DocumentationHandler) getGoDocumentation(symbol string) (*DocResponse, error) {
	// Common Go packages, functions, and methods
	docs := map[string]DocResponse{
		"fmt.Println": {
			Symbol:      "fmt.Println",
			Description: "Prints to standard output and appends a newline.",
			Syntax:      "func Println(a ...interface{}) (n int, err error)",
			Example:     "fmt.Println(\"Hello, World!\")",
			URL:         "https://pkg.go.dev/fmt#Println",
		},
		"http.ListenAndServe": {
			Symbol:      "http.ListenAndServe",
			Description: "Starts an HTTP server with a given address and handler.",
			Syntax:      "func ListenAndServe(addr string, handler Handler) error",
			Example:     "http.ListenAndServe(\":8080\", nil)",
			URL:         "https://pkg.go.dev/net/http#ListenAndServe",
		},
		"json.Marshal": {
			Symbol:      "json.Marshal",
			Description: "Returns the JSON encoding of v.",
			Syntax:      "func Marshal(v interface{}) ([]byte, error)",
			Example:     "data, err := json.Marshal(myStruct)",
			URL:         "https://pkg.go.dev/encoding/json#Marshal",
		},
		"struct": {
			Symbol:      "struct",
			Description: "A struct is a sequence of named elements, called fields, each of which has a name and a type.",
			Syntax:      "type StructName struct {\n\tField1 Type1\n\tField2 Type2\n}",
			Example:     "type Person struct {\n\tName string\n\tAge int\n}",
			URL:         "https://go.dev/ref/spec#Struct_types",
		},
		"interface": {
			Symbol:      "interface",
			Description: "An interface type specifies a method set called its interface.",
			Syntax:      "type InterfaceName interface {\n\tMethod1() ReturnType\n\tMethod2(Type) ReturnType\n}",
			Example:     "type Writer interface {\n\tWrite(p []byte) (n int, err error)\n}",
			URL:         "https://go.dev/ref/spec#Interface_types",
		},
		"goroutine": {
			Symbol:      "goroutine",
			Description: "Goroutines are lightweight threads managed by the Go runtime.",
			Syntax:      "go functionCall()",
			Example:     "go func() {\n\tfmt.Println(\"Hello from goroutine\")\n}()",
			URL:         "https://go.dev/doc/effective_go#goroutines",
		},
		"channel": {
			Symbol:      "channel",
			Description: "Channels are typed conduits through which you can send and receive values with the channel operator, <-.",
			Syntax:      "ch := make(chan Type)\nch <- v    // Send v to channel ch\nv := <-ch  // Receive from ch, and assign value to v",
			Example:     "ch := make(chan int)\ngo func() { ch <- 42 }()\nfmt.Println(<-ch)",
			URL:         "https://go.dev/doc/effective_go#channels",
		},
	}

	// Check if we have documentation for the symbol
	if doc, ok := docs[symbol]; ok {
		return &doc, nil
	}

	// Handle partial matches
	for key, doc := range docs {
		if strings.Contains(key, symbol) {
			return &doc, nil
		}
	}

	// Handle Go keywords
	keywords := map[string]DocResponse{
		"if": {
			Symbol:      "if",
			Description: "Conditional statement that executes code based on the evaluation of a condition.",
			Syntax:      "if condition {\n\t// code\n} else if condition {\n\t// code\n} else {\n\t// code\n}",
			Example:     "if x > 0 {\n\tfmt.Println(\"Positive\")\n} else if x < 0 {\n\tfmt.Println(\"Negative\")\n} else {\n\tfmt.Println(\"Zero\")\n}",
			URL:         "https://go.dev/ref/spec#If_statements",
		},
		"for": {
			Symbol:      "for",
			Description: "Loop that iterates while a condition is true, or iterates over a range.",
			Syntax:      "for initialization; condition; post {\n\t// code\n}\n\nfor condition {\n\t// code\n}\n\nfor range expression {\n\t// code\n}",
			Example:     "for i := 0; i < 10; i++ {\n\tfmt.Println(i)\n}",
			URL:         "https://go.dev/ref/spec#For_statements",
		},
		"switch": {
			Symbol:      "switch",
			Description: "Conditional statement that evaluates an expression and executes the matching case.",
			Syntax:      "switch expression {\ncase value1:\n\t// code\ncase value2:\n\t// code\ndefault:\n\t// code\n}",
			Example:     "switch day {\ncase \"Monday\":\n\tfmt.Println(\"Start of work week\")\ncase \"Friday\":\n\tfmt.Println(\"End of work week\")\ndefault:\n\tfmt.Println(\"Regular day\")\n}",
			URL:         "https://go.dev/ref/spec#Switch_statements",
		},
	}

	// Check if we have documentation for the keyword
	if doc, ok := keywords[symbol]; ok {
		return &doc, nil
	}

	// Return generic documentation if no specific documentation is found
	return &DocResponse{
		Symbol:      symbol,
		Description: fmt.Sprintf("Go symbol: %s", symbol),
		URL:         fmt.Sprintf("https://pkg.go.dev/search?q=%s", symbol),
	}, nil
}

// getPythonDocumentation returns Python-specific documentation
func (h *DocumentationHandler) getPythonDocumentation(symbol string) (*DocResponse, error) {
	// Common Python functions, methods, and modules
	docs := map[string]DocResponse{
		"print": {
			Symbol:      "print",
			Description: "Prints the specified message to the screen, or other standard output device.",
			Syntax:      "print(*objects, sep=' ', end='\\n', file=sys.stdout, flush=False)",
			Example:     "print(\"Hello, World!\")",
			URL:         "https://docs.python.org/3/library/functions.html#print",
		},
		"len": {
			Symbol:      "len",
			Description: "Returns the number of items in an object.",
			Syntax:      "len(s)",
			Example:     "length = len([1, 2, 3])",
			URL:         "https://docs.python.org/3/library/functions.html#len",
		},
		"list": {
			Symbol:      "list",
			Description: "Creates a list object or converts an iterable to a list.",
			Syntax:      "list([iterable])",
			Example:     "my_list = list(range(5))",
			URL:         "https://docs.python.org/3/library/functions.html#func-list",
		},
		"dict": {
			Symbol:      "dict",
			Description: "Creates a dictionary object or converts mapping/iterable to a dictionary.",
			Syntax:      "dict(**kwargs)\ndict(mapping, **kwargs)\ndict(iterable, **kwargs)",
			Example:     "person = dict(name=\"John\", age=30)",
			URL:         "https://docs.python.org/3/library/functions.html#func-dict",
		},
		"open": {
			Symbol:      "open",
			Description: "Opens a file and returns a file object.",
			Syntax:      "open(file, mode='r', buffering=-1, encoding=None, errors=None, newline=None, closefd=True, opener=None)",
			Example:     "with open('file.txt', 'r') as f:\n    content = f.read()",
			URL:         "https://docs.python.org/3/library/functions.html#open",
		},
		"range": {
			Symbol:      "range",
			Description: "Returns a sequence of numbers, starting from 0 by default, and increments by 1 by default, and stops before a specified number.",
			Syntax:      "range(stop)\nrange(start, stop[, step])",
			Example:     "for i in range(5):\n    print(i)",
			URL:         "https://docs.python.org/3/library/functions.html#func-range",
		},
	}

	// Check if we have documentation for the symbol
	if doc, ok := docs[symbol]; ok {
		return &doc, nil
	}

	// Handle partial matches
	for key, doc := range docs {
		if strings.Contains(key, symbol) {
			return &doc, nil
		}
	}

	// Return generic documentation if no specific documentation is found
	return &DocResponse{
		Symbol:      symbol,
		Description: fmt.Sprintf("Python symbol: %s", symbol),
		URL:         fmt.Sprintf("https://docs.python.org/3/search.html?q=%s", symbol),
	}, nil
}

// getSQLDocumentation returns SQL-specific documentation
func (h *DocumentationHandler) getSQLDocumentation(symbol string) (*DocResponse, error) {
	// Common SQL commands and functions
	docs := map[string]DocResponse{
		"SELECT": {
			Symbol:      "SELECT",
			Description: "Extracts data from a database.",
			Syntax:      "SELECT column1, column2, ...\nFROM table_name\nWHERE condition\nGROUP BY column\nHAVING condition\nORDER BY column;",
			Example:     "SELECT name, age FROM users WHERE age > 18 ORDER BY name;",
			URL:         "https://www.w3schools.com/sql/sql_select.asp",
		},
		"INSERT": {
			Symbol:      "INSERT",
			Description: "Inserts new data into a database.",
			Syntax:      "INSERT INTO table_name (column1, column2, ...)\nVALUES (value1, value2, ...);",
			Example:     "INSERT INTO users (name, age) VALUES ('John', 25);",
			URL:         "https://www.w3schools.com/sql/sql_insert.asp",
		},
		"UPDATE": {
			Symbol:      "UPDATE",
			Description: "Updates existing data in a database.",
			Syntax:      "UPDATE table_name\nSET column1 = value1, column2 = value2, ...\nWHERE condition;",
			Example:     "UPDATE users SET age = 26 WHERE name = 'John';",
			URL:         "https://www.w3schools.com/sql/sql_update.asp",
		},
		"DELETE": {
			Symbol:      "DELETE",
			Description: "Deletes data from a database.",
			Syntax:      "DELETE FROM table_name WHERE condition;",
			Example:     "DELETE FROM users WHERE name = 'John';",
			URL:         "https://www.w3schools.com/sql/sql_delete.asp",
		},
		"JOIN": {
			Symbol:      "JOIN",
			Description: "Combines rows from two or more tables, based on a related column between them.",
			Syntax:      "SELECT columns\nFROM table1\nJOIN table2\nON table1.column = table2.column;",
			Example:     "SELECT users.name, orders.product FROM users JOIN orders ON users.id = orders.user_id;",
			URL:         "https://www.w3schools.com/sql/sql_join.asp",
		},
	}

	// Convert symbol to uppercase for SQL commands
	upperSymbol := strings.ToUpper(symbol)

	// Check if we have documentation for the symbol
	if doc, ok := docs[upperSymbol]; ok {
		return &doc, nil
	}

	// Handle partial matches
	for key, doc := range docs {
		if strings.Contains(key, upperSymbol) {
			return &doc, nil
		}
	}

	// Return generic documentation if no specific documentation is found
	return &DocResponse{
		Symbol:      symbol,
		Description: fmt.Sprintf("SQL command or function: %s", symbol),
		URL:         fmt.Sprintf("https://www.w3schools.com/sql/sql_%s.asp", strings.ToLower(symbol)),
	}, nil
}

// getGenericDocumentation returns generic documentation for unsupported languages
func (h *DocumentationHandler) getGenericDocumentation(symbol string, language string) (*DocResponse, error) {
	return &DocResponse{
		Symbol:      symbol,
		Description: fmt.Sprintf("%s symbol: %s", language, symbol),
		URL:         fmt.Sprintf("https://www.google.com/search?q=%s+%s+documentation", language, symbol),
	}, nil
}

// ToJSON converts the documentation response to JSON
func (r *DocResponse) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
