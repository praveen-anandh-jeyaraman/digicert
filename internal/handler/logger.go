package handler

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "time"
)

// StructuredLogger handles structured logging with request context
type StructuredLogger struct {
    logger *log.Logger
}

type LogEntry struct {
    Timestamp string                 `json:"timestamp"`
    RequestID string                 `json:"request_id"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger() *StructuredLogger {
    return &StructuredLogger{
        logger: log.New(os.Stdout, "", 0),
    }
}

// Info logs info level message
func (sl *StructuredLogger) Info(ctx context.Context, message string, data map[string]interface{}) {
    sl.log(ctx, "INFO", message, data)
}

// Error logs error level message
func (sl *StructuredLogger) Error(ctx context.Context, message string, data map[string]interface{}) {
    sl.log(ctx, "ERROR", message, data)
}

// Warn logs warning level message
func (sl *StructuredLogger) Warn(ctx context.Context, message string, data map[string]interface{}) {
    sl.log(ctx, "WARN", message, data)
}

// Debug logs debug level message
func (sl *StructuredLogger) Debug(ctx context.Context, message string, data map[string]interface{}) {
    sl.log(ctx, "DEBUG", message, data)
}

func (sl *StructuredLogger) log(ctx context.Context, level, message string, data map[string]interface{}) {
    entry := LogEntry{
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: GetRequestID(ctx),
        Level:     level,
        Message:   message,
        Data:      data,
    }

    jsonBytes, _ := json.Marshal(entry)
    sl.logger.Println(string(jsonBytes))
}