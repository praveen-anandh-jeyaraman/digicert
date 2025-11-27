package handler

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/logger"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
)

type AuthHandler struct {
    authSvc service.AuthService
    userSvc service.UserService
}

func NewAuthHandler(authSvc service.AuthService, userSvc service.UserService) *AuthHandler {
    return &AuthHandler{
        authSvc: authSvc,
        userSvc: userSvc,
    }
}

// Login godoc
// @Summary      Login user
// @Description  Login with username and password
// @Tags         Auth
// @Accept       json
// @Param        request  body      model.LoginRequest  true  "Login credentials"
// @Produce      json
// @Success      200  {object}  model.LoginResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    var req model.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    user, err := h.userSvc.ValidatePassword(r.Context(), req.Username, req.Password)
    if err != nil {
        log.Printf("[%s] Login failed: %v", requestID, err)

        // Track failed login
        cwLogger := logger.GetLogger()
        if cwLogger != nil {
            cwLogger.PutMetric(r.Context(), "LoginFailed", 1, "Count")
        }
        WriteError(r.Context(), w, http.StatusUnauthorized, "Invalid username or password")
        return
    }

    token, expiresAt, err := h.authSvc.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        log.Printf("[%s] Token generation failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to generate token")
        return
    }

    resp := model.LoginResponse{
        Token:     token,
        ExpiresAt: expiresAt,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(resp)
    log.Printf("[%s] User logged in: %s (role: %s)", requestID, user.Username, user.Role)
}

// Refresh godoc
// @Summary      Refresh token
// @Description  Get a new token
// @Tags         Auth
// @Accept       json
// @Param        request  body      model.RefreshRequest  true  "Current token"
// @Produce      json
// @Success      200  {object}  model.LoginResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    var req model.RefreshRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    claims, err := h.authSvc.ValidateToken(req.Token)
    if err != nil {
        log.Printf("[%s] Token validation failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Invalid token")
        return
    }

    userID := claims["user_id"].(string)
    username := claims["username"].(string)
    role := claims["role"].(string)

    token, expiresAt, err := h.authSvc.GenerateToken(userID, username, role)
    if err != nil {
        log.Printf("[%s] Token generation failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to generate token")
        return
    }

    resp := model.LoginResponse{
        Token:     token,
        ExpiresAt: expiresAt,
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
    log.Printf("[%s] Token refreshed for user: %s", requestID, username)
}