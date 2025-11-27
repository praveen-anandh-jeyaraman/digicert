package handler

import (
    "context"
    "log"
    "net/http"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
)

// GetRole retrieves role from context
func GetRole(r *http.Request) string {
    role, ok := r.Context().Value("role").(string)
    if !ok {
        return ""
    }
    return role
}

// AdminMiddleware checks if user is admin
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := GetRequestID(r.Context())

        // Get role from context (set by AuthMiddleware)
        role, ok := r.Context().Value("role").(string)
        if !ok || role != "admin" {
            log.Printf("[%s] Admin access denied. Role: %v", requestID, role)
            WriteError(r.Context(), w, http.StatusForbidden, "Admin access required")
            return
        }

        next.ServeHTTP(w, r)
    })
}

// AuthMiddleware checks JWT and extracts user info + role
func AuthMiddleware(authSvc service.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := GetRequestID(r.Context())

            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                log.Printf("[%s] Missing authorization header", requestID)
                WriteError(r.Context(), w, http.StatusUnauthorized, "Missing authorization header")
                return
            }

            token := authHeader[7:]
            claims, err := authSvc.ValidateToken(token)
            if err != nil {
                log.Printf("[%s] Invalid token: %v", requestID, err)
                WriteError(r.Context(), w, http.StatusUnauthorized, "Invalid token")
                return
            }

            // Add user info to context
            ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
            ctx = context.WithValue(ctx, "username", claims["username"])
            ctx = context.WithValue(ctx, "role", claims["role"])

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}