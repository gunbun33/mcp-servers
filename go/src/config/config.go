package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the MCP server
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	MCP        MCPConfig        `mapstructure:"mcp"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Database   DatabaseConfig   `mapstructure:"database"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        int      `mapstructure:"port"`
	Host        string   `mapstructure:"host"`
	MetricsPort int      `mapstructure:"metrics_port"`
	LogLevel    string   `mapstructure:"log_level"`
	Debug       bool     `mapstructure:"debug"`
	CORS        CORSConfig `mapstructure:"cors"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// MCPConfig holds MCP protocol configuration
type MCPConfig struct {
	ProtocolVersion string            `mapstructure:"protocol_version"`
	ServerName      string            `mapstructure:"server_name"`
	ServerVersion   string            `mapstructure:"server_version"`
	Capabilities    CapabilitiesConfig `mapstructure:"capabilities"`
}

// CapabilitiesConfig holds MCP capabilities configuration
type CapabilitiesConfig struct {
	SupportedLanguages []string `mapstructure:"supported_languages"`
	SupportsNotebooks  bool     `mapstructure:"supports_notebooks"`
	SupportsStreaming  bool     `mapstructure:"supports_streaming"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Format   string         `mapstructure:"format"`
	Output   string         `mapstructure:"output"`
	File     string         `mapstructure:"file"`
	Rotation RotationConfig `mapstructure:"rotation"`
}

// RotationConfig holds log rotation configuration
type RotationConfig struct {
	MaxSize    int `mapstructure:"max_size"`
	MaxBackups int `mapstructure:"max_backups"`
	MaxAge     int `mapstructure:"max_age"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Prometheus          bool   `mapstructure:"prometheus"`
	HealthCheckInterval string `mapstructure:"health_check_interval"`
	HealthCheckTimeout  string `mapstructure:"health_check_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
	MaxConnections   int    `mapstructure:"max_connections"`
	ConnectionTimeout string `mapstructure:"connection_timeout"`
	IdleTimeout      string `mapstructure:"idle_timeout"`
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	overrideWithEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 9090)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.metrics_port", 9091)
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("server.debug", false)
	viper.SetDefault("server.cors.allowed_origins", []string{"*"})
	viper.SetDefault("server.cors.allowed_methods", []string{"GET", "POST", "OPTIONS"})
	viper.SetDefault("server.cors.allowed_headers", []string{"*"})

	// MCP defaults
	viper.SetDefault("mcp.protocol_version", "2.0")
	viper.SetDefault("mcp.server_name", "Go MCP Server")
	viper.SetDefault("mcp.server_version", "1.0.0")
	viper.SetDefault("mcp.capabilities.supported_languages", []string{"go", "sql"})
	viper.SetDefault("mcp.capabilities.supports_notebooks", true)
	viper.SetDefault("mcp.capabilities.supports_streaming", true)

	// Logging defaults
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.file", "/var/log/mcp-go-agent.log")
	viper.SetDefault("logging.rotation.max_size", 10)
	viper.SetDefault("logging.rotation.max_backups", 3)
	viper.SetDefault("logging.rotation.max_age", 7)

	// Monitoring defaults
	viper.SetDefault("monitoring.prometheus", true)
	viper.SetDefault("monitoring.health_check_interval", "30s")
	viper.SetDefault("monitoring.health_check_timeout", "10s")

	// Database defaults
	viper.SetDefault("database.connection_string", "postgres://demo:demo@db:5432/demo?sslmode=disable")
	viper.SetDefault("database.max_connections", 10)
	viper.SetDefault("database.connection_timeout", "5s")
	viper.SetDefault("database.idle_timeout", "60s")
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv() {
	// Server environment variables
	if port := os.Getenv("PORT"); port != "" {
		viper.Set("server.port", port)
	}
	if host := os.Getenv("HOST"); host != "" {
		viper.Set("server.host", host)
	}
	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		viper.Set("server.metrics_port", metricsPort)
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		viper.Set("server.log_level", logLevel)
	}
	if debug := os.Getenv("DEBUG"); debug != "" {
		viper.Set("server.debug", debug == "true")
	}

	// MCP environment variables
	if serverName := os.Getenv("MCP_SERVER_NAME"); serverName != "" {
		viper.Set("mcp.server_name", serverName)
	}
	if serverVersion := os.Getenv("MCP_SERVER_VERSION"); serverVersion != "" {
		viper.Set("mcp.server_version", serverVersion)
	}

	// Database environment variables
	if connString := os.Getenv("DATABASE_URL"); connString != "" {
		viper.Set("database.connection_string", connString)
	}
}
