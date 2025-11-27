package service

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
    GenerateToken(userID, username, role string) (string, time.Time, error)
    ValidateToken(token string) (map[string]interface{}, error)
}

type authService struct {
    secretKey string
    expiry    time.Duration
}

func NewAuthService(secretKey string, expiry time.Duration) AuthService {
    return &authService{
        secretKey: secretKey,
        expiry:    expiry,
    }
}

type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func (s *authService) GenerateToken(userID, username, role string) (string, time.Time, error) {
    expiresAt := time.Now().Add(s.expiry)
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiresAt),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(s.secretKey))
    if err != nil {
        return "", time.Time{}, err
    }

    return tokenString, expiresAt, nil
}

func (s *authService) ValidateToken(tokenString string) (map[string]interface{}, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.secretKey), nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }

    return map[string]interface{}{
        "user_id":  claims.UserID,
        "username": claims.Username,
        "role":     claims.Role,
    }, nil
}