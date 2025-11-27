package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/stretchr/testify/require"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// Mock auth service
type mockAuthService struct {
    generateFn func(userID, username, role string) (string, time.Time, error)
    validateFn func(token string) (map[string]interface{}, error)
}

func (m *mockAuthService) GenerateToken(userID, username, role string) (string, time.Time, error) {
    return m.generateFn(userID, username, role)
}

func (m *mockAuthService) ValidateToken(token string) (map[string]interface{}, error) {
    return m.validateFn(token)
}

// Mock user service
type mockUserServiceForAuth struct {
    registerFn      func(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
    getByIDFn       func(ctx context.Context, id string) (*model.User, error)
    updateFn        func(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
    validateFn      func(ctx context.Context, username, password string) (*model.User, error)
    getByEmailFn    func(ctx context.Context, email string) (*model.User, error)
    getByUsernameFn func(ctx context.Context, username string) (*model.User, error)
    listFn          func(ctx context.Context, limit, offset int) ([]model.User, error)
    deleteFn        func(ctx context.Context, id string) error
}

func (m *mockUserServiceForAuth) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
    return m.registerFn(ctx, req)
}

func (m *mockUserServiceForAuth) GetByID(ctx context.Context, id string) (*model.User, error) {
    return m.getByIDFn(ctx, id)
}

func (m *mockUserServiceForAuth) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
    return m.updateFn(ctx, id, updates)
}

func (m *mockUserServiceForAuth) ValidatePassword(ctx context.Context, username, password string) (*model.User, error) {
    return m.validateFn(ctx, username, password)
}

func (m *mockUserServiceForAuth) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    return m.getByEmailFn(ctx, email)
}

func (m *mockUserServiceForAuth) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return m.getByUsernameFn(ctx, username)
}

func (m *mockUserServiceForAuth) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    return m.listFn(ctx, limit, offset)
}

func (m *mockUserServiceForAuth) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

// Helper to set request ID in context properly
func createAuthRequest(method, path string, body string, requestID string) *http.Request {
    req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := context.WithValue(req.Context(), RequestIDKey, requestID)
    return req.WithContext(ctx)
}

func TestAuthHandler_Login_Success(t *testing.T) {
    mockAuthSvc := &mockAuthService{
        generateFn: func(userID, username, role string) (string, time.Time, error) {
            return "valid-token", time.Now().Add(24 * time.Hour), nil
        },
    }
    mockUserSvc := &mockUserServiceForAuth{
        validateFn: func(_ context.Context, username, password string) (*model.User, error) {
            return &model.User{
                ID:       "user-1",
                Username: username,
                Role:     "USER",
            }, nil
        },
    }
    h := NewAuthHandler(mockAuthSvc, mockUserSvc)

    req := createAuthRequest("POST", "/auth/login", `{"username":"john","password":"SecurePass123"}`, "test-auth-001")
    rec := httptest.NewRecorder()

    h.Login(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var resp model.LoginResponse
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
    require.NotEmpty(t, resp.Token)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
    mockAuthSvc := &mockAuthService{}
    mockUserSvc := &mockUserServiceForAuth{
        validateFn: func(_ context.Context, username, password string) (*model.User, error) {
            return nil, ErrInvalidCredentials
        },
    }
    h := NewAuthHandler(mockAuthSvc, mockUserSvc)

    req := createAuthRequest("POST", "/auth/login", `{"username":"john","password":"WrongPassword"}`, "test-auth-002")
    rec := httptest.NewRecorder()

    h.Login(rec, req)
    require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
    mockAuthSvc := &mockAuthService{
        validateFn: func(token string) (map[string]interface{}, error) {
            return map[string]interface{}{
                "user_id":  "user-1",
                "username": "john",
                "role":     "USER",
            }, nil
        },
        generateFn: func(userID, username, role string) (string, time.Time, error) {
            return "new-token", time.Now().Add(24 * time.Hour), nil
        },
    }
    mockUserSvc := &mockUserServiceForAuth{}
    h := NewAuthHandler(mockAuthSvc, mockUserSvc)

    req := createAuthRequest("POST", "/auth/refresh", `{"token":"old-token"}`, "test-auth-003")
    rec := httptest.NewRecorder()

    h.Refresh(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var resp model.LoginResponse
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
    require.Equal(t, "new-token", resp.Token)
}