package handler

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
)

// ErrorResponse is a standard error format
type ErrorResponse struct {
    RequestID string `json:"request_id"`
    Error     string `json:"error"`
    Message   string `json:"message,omitempty"`
    Status    int    `json:"status"`
}

// WriteError writes a standardized error response with request ID
func WriteError(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)

    requestID := GetRequestID(ctx)
    resp := ErrorResponse{
        RequestID: requestID,
        Error:     http.StatusText(statusCode),
        Message:   message,
        Status:    statusCode,
    }

    if err := json.NewEncoder(w).Encode(resp); err != nil {
        log.Printf("[%s] failed to encode error response: %v", requestID, err)
    }
}

// WriteValidationErrors writes validation errors with request ID
func WriteValidationErrors(ctx context.Context, w http.ResponseWriter, errs ValidationErrors) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)

    requestID := GetRequestID(ctx)
    response := map[string]interface{}{
        "request_id": requestID,
        "errors":     errs,
    }

    _ = json.NewEncoder(w).Encode(response)
}