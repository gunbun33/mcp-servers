package logger

import (
	"io"
	"os"

	"github.com/cploutarchou/mcp-servers/go/config"
	"github.com/sirupsen/logrus"
)

// Logger is a wrapper around logrus.Logger
type Logger struct {
	*logrus.Logger
}

// New creates a new logger
func New(cfg *config.LoggingConfig) *Logger {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Format)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set log output
	var output io.Writer
	if cfg.Output == "stdout" {
		output = os.Stdout
	} else if cfg.Output == "stderr" {
		output = os.Stderr
	} else if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			output = file
		} else {
			log.Warnf("Failed to log to file %s, using stdout instead: %v", cfg.File, err)
			output = os.Stdout
		}
	} else {
		output = os.Stdout
	}

	log.SetOutput(output)

	return &Logger{
		Logger: log,
	}
}
