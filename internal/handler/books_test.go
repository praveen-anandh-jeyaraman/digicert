package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/stretchr/testify/require"
)

// Helper to set request ID in context properly
func createTestRequest(method, path string, body string, requestID string) *http.Request {
    req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := context.WithValue(req.Context(), RequestIDKey, requestID)
    return req.WithContext(ctx)
}

// Mock user service for book handler tests
type mockUserServiceForBooks struct {
    registerFn      func(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
    getByIDFn       func(ctx context.Context, id string) (*model.User, error)
    updateFn        func(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
    validateFn      func(ctx context.Context, username, password string) (*model.User, error)
    getByEmailFn    func(ctx context.Context, email string) (*model.User, error)
    getByUsernameFn func(ctx context.Context, username string) (*model.User, error)
    listFn          func(ctx context.Context, limit, offset int) ([]model.User, error)
    deleteFn        func(ctx context.Context, id string) error
}

func (m *mockUserServiceForBooks) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
    return m.registerFn(ctx, req)
}

func (m *mockUserServiceForBooks) GetByID(ctx context.Context, id string) (*model.User, error) {
    return m.getByIDFn(ctx, id)
}

func (m *mockUserServiceForBooks) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
    return m.updateFn(ctx, id, updates)
}

func (m *mockUserServiceForBooks) ValidatePassword(ctx context.Context, username, password string) (*model.User, error) {
    return m.validateFn(ctx, username, password)
}

func (m *mockUserServiceForBooks) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    return m.getByEmailFn(ctx, email)
}

func (m *mockUserServiceForBooks) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return m.getByUsernameFn(ctx, username)
}

func (m *mockUserServiceForBooks) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    return m.listFn(ctx, limit, offset)
}

func (m *mockUserServiceForBooks) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

// Mock book service
type mockBookServiceForHandler struct {
    listFn    func(ctx context.Context, limit, offset int) ([]model.Book, error)
    getByIDFn func(ctx context.Context, id string) (model.Book, error)
    createFn  func(ctx context.Context, b *model.Book) error
    updateFn  func(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error)
    deleteFn  func(ctx context.Context, id string) error
}

func (m *mockBookServiceForHandler) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
    return m.listFn(ctx, limit, offset)
}

func (m *mockBookServiceForHandler) GetByID(ctx context.Context, id string) (model.Book, error) {
    return m.getByIDFn(ctx, id)
}

func (m *mockBookServiceForHandler) Create(ctx context.Context, b *model.Book) error {
    if m.createFn == nil {
        return errors.New("createFn not set")
    }
    return m.createFn(ctx, b)
}

func (m *mockBookServiceForHandler) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
    return m.updateFn(ctx, id, updates)
}

func (m *mockBookServiceForHandler) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

// User Handler Tests

func TestUserHandler_Register_Success(t *testing.T) {
    mock := &mockUserServiceForBooks{
        registerFn: func(_ context.Context, req *model.RegisterRequest) (*model.User, error) {
            user := &model.User{
                ID:       "user-1",
                Username: req.Username,
                Email:    req.Email,
                Role:     "USER",
            }
            return user, nil
        },
    }
    h := NewUserHandler(mock)

    req := createTestRequest("POST", "/auth/register", `{"username":"john","email":"john@example.com","password":"SecurePass123"}`, "test-user-001")
    rec := httptest.NewRecorder()

    h.Register(rec, req)
    require.Equal(t, http.StatusCreated, rec.Code)

    var user model.User
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &user))
    require.Equal(t, "john", user.Username)
    require.Equal(t, "john@example.com", user.Email)
    require.NotEmpty(t, user.ID)
}

func TestUserHandler_Register_InvalidEmail(t *testing.T) {
    mock := &mockUserServiceForBooks{}
    h := NewUserHandler(mock)

    req := createTestRequest("POST", "/auth/register", `{"username":"john","email":"invalid-email","password":"SecurePass123"}`, "test-user-002")
    rec := httptest.NewRecorder()

    h.Register(rec, req)
    require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
    mock := &mockUserServiceForBooks{
        getByIDFn: func(_ context.Context, id string) (*model.User, error) {
            return &model.User{
                ID:       id,
                Username: "john",
                Email:    "john@example.com",
                Role:     "USER",
            }, nil
        },
    }
    h := NewUserHandler(mock)

    req := createTestRequest("GET", "/users/me", "", "test-user-003")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "user_id", "user-1")
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.GetProfile(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var user model.User
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &user))
    require.Equal(t, "john", user.Username)
}

func TestUserHandler_ListUsers_Success(t *testing.T) {
    mock := &mockUserServiceForBooks{
        listFn: func(_ context.Context, limit, offset int) ([]model.User, error) {
            return []model.User{
                {ID: "1", Username: "john", Role: "USER"},
                {ID: "2", Username: "admin", Role: "ADMIN"},
            }, nil
        },
    }
    h := NewUserHandler(mock)

    req := createTestRequest("GET", "/admin/users", "", "test-user-004")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "role", "ADMIN")
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.ListUsers(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var users []model.User
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &users))
    require.Len(t, users, 2)
}

// Book Handler Tests

func TestBookHandler_List_Success(t *testing.T) {
    svc := &mockBookServiceForHandler{
        listFn: func(_ context.Context, limit, offset int) ([]model.Book, error) {
            return []model.Book{
                {ID: "1", Title: "Test Book", Author: "Test Author"},
            }, nil
        },
    }

    h := NewBookHandler(svc)

    req := createTestRequest("GET", "/books?limit=10&offset=0", "", "test-book-001")
    rec := httptest.NewRecorder()

    h.List(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var books []model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &books))
    require.NotEmpty(t, books)
}

func TestBookHandler_Get_Success(t *testing.T) {
    svc := &mockBookServiceForHandler{
        getByIDFn: func(_ context.Context, id string) (model.Book, error) {
            return model.Book{ID: "1", Title: "Test Book", Author: "Test Author"}, nil
        },
    }

    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "1")
    req := createTestRequest("GET", "/books/1", "", "test-book-002")
    ctx := req.Context()
    ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Get(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var book model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &book))
    require.Equal(t, "1", book.ID)
}

func TestBookHandler_Get_NotFound(t *testing.T) {
    svc := &mockBookServiceForHandler{
        getByIDFn: func(_ context.Context, id string) (model.Book, error) {
            return model.Book{}, errors.New("book not found")
        },
    }

    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "nonexistent")
    req := createTestRequest("GET", "/books/nonexistent", "", "test-book-003")
    ctx := req.Context()
    ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Get(rec, req)
    require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBookHandler_Create_Success(t *testing.T) {
    svc := &mockBookServiceForHandler{
        createFn: func(_ context.Context, b *model.Book) error {
            b.ID = "test-book-1"
            return nil
        },
    }
    h := NewBookHandler(svc)

    req := createTestRequest("POST", "/books", `{"title":"Go Programming","author":"John Doe","published_year":2020}`, "test-book-004")
    rec := httptest.NewRecorder()

    h.Create(rec, req)
    require.Equal(t, http.StatusCreated, rec.Code)

    var created model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
    require.Equal(t, "Go Programming", created.Title)
    require.Equal(t, "John Doe", created.Author)
}

func TestBookHandler_Create_ServiceError(t *testing.T) {
    svc := &mockBookServiceForHandler{
        createFn: func(_ context.Context, b *model.Book) error {
            return errors.New("service error")
        },
    }
    h := NewBookHandler(svc)

    req := createTestRequest("POST", "/books", `{"title":"Go Programming","author":"John Doe","published_year":2020}`, "test-book-005")
    rec := httptest.NewRecorder()

    h.Create(rec, req)
    require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestBookHandler_Update_Success(t *testing.T) {
    svc := &mockBookServiceForHandler{
        updateFn: func(_ context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
            return &model.Book{
                ID:     id,
                Title:  "Updated Title",
                Author: "Updated Author",
            }, nil
        },
    }
    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "1")
    req := createTestRequest("PUT", "/books/1", `{"title":"Updated Title","author":"Updated Author"}`, "test-book-006")
    ctx := req.Context()
    ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Update(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var updated model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
    require.Equal(t, "Updated Title", updated.Title)
}

func TestBookHandler_Delete_Success(t *testing.T) {
    svc := &mockBookServiceForHandler{
        deleteFn: func(_ context.Context, id string) error {
            return nil
        },
    }
    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "1")
    req := createTestRequest("DELETE", "/books/1", "", "test-book-007")
    ctx := req.Context()
    ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Delete(rec, req)
    require.Equal(t, http.StatusNoContent, rec.Code)
}