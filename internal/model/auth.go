package model

import "time"

type LoginResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}

type RefreshRequest struct {
    Token string `json:"token"`
}

type Claims struct {
    Username string `json:"username"`
    UserID   string `json:"user_id"`
    // Standard JWT claims
    ExpiresAt int64 `json:"exp"`
    IssuedAt  int64 `json:"iat"`
}