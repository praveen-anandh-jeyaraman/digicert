package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
    "github.com/stretchr/testify/require"
)

// Helper to add request ID to context
func createRequestWithID(method, path string, body *bytes.Buffer) *http.Request {
    var req *http.Request
    if body != nil {
        req = httptest.NewRequest(method, path, body)
    } else {
        req = httptest.NewRequest(method, path, nil)
    }

    // Add request ID to context
    ctx := context.WithValue(req.Context(), RequestIDKey, "test-req-123")
    return req.WithContext(ctx)
}

// Mock service for testing
type mockBookService struct {
    books map[string]*model.Book
}

func (m *mockBookService) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
    books := make([]model.Book, 0)
    for _, b := range m.books {
        books = append(books, *b)
    }
    return books, nil
}

func (m *mockBookService) Get(ctx context.Context, id string) (model.Book, error) {
    if b, ok := m.books[id]; ok {
        return *b, nil
    }
    return model.Book{}, service.ErrConflict
}

func (m *mockBookService) Create(ctx context.Context, b *model.Book) error {
    b.ID = "test-book-1"
    m.books[b.ID] = b
    return nil
}

func (m *mockBookService) Update(ctx context.Context, b *model.Book) error {
    if _, ok := m.books[b.ID]; !ok {
        return service.ErrConflict
    }
    m.books[b.ID] = b
    return nil
}

func (m *mockBookService) Delete(ctx context.Context, id string) error {
    if _, ok := m.books[id]; !ok {
        return service.ErrConflict
    }
    delete(m.books, id)
    return nil
}

func newMockBookService() *mockBookService {
    return &mockBookService{books: make(map[string]*model.Book)}
}

// Tests

func TestBookHandler_List_Success(t *testing.T) {
    svc := newMockBookService()
    svc.books["1"] = &model.Book{ID: "1", Title: "Test Book", Author: "Test Author"}

    h := NewBookHandler(svc)

    req := createRequestWithID("GET", "/books?limit=10&offset=0", nil)
    rec := httptest.NewRecorder()

    h.List(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var books []model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &books))
    require.NotEmpty(t, books)
}

func TestBookHandler_List_InvalidLimit(t *testing.T) {
    svc := newMockBookService()
    h := NewBookHandler(svc)

    req := createRequestWithID("GET", "/books?limit=999", nil)
    rec := httptest.NewRecorder()

    h.List(rec, req)
    require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBookHandler_Get_Success(t *testing.T) {
    svc := newMockBookService()
    svc.books["1"] = &model.Book{ID: "1", Title: "Test Book", Author: "Test Author"}

    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "1")
    req := createRequestWithID("GET", "/books/1", nil)
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
    rec := httptest.NewRecorder()

    h.Get(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var book model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &book))
    require.Equal(t, "1", book.ID)
}

func TestBookHandler_Get_NotFound(t *testing.T) {
    svc := newMockBookService()
    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "nonexistent")
    req := createRequestWithID("GET", "/books/nonexistent", nil)
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
    rec := httptest.NewRecorder()

    h.Get(rec, req)
    require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBookHandler_Create_Success(t *testing.T) {
    svc := newMockBookService()
    h := NewBookHandler(svc)

    createBody := `{"title":"Go Programming","author":"John Doe","published_year":2020}`
    req := createRequestWithID("POST", "/books", bytes.NewBufferString(createBody))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    h.Create(rec, req)
    require.Equal(t, http.StatusCreated, rec.Code)

    var created model.Book
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
    require.Equal(t, "Go Programming", created.Title)
    require.Equal(t, "John Doe", created.Author)
}

func TestBookHandler_Create_BadRequest(t *testing.T) {
    svc := newMockBookService()
    h := NewBookHandler(svc)

    createBody := `{"title":"","author":""}`
    req := createRequestWithID("POST", "/books", bytes.NewBufferString(createBody))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    h.Create(rec, req)
    require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBookHandler_Update_Success(t *testing.T) {
    svc := newMockBookService()
    svc.books["1"] = &model.Book{ID: "1", Title: "Old Title", Author: "Old Author"}

    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "1")
    updateBody := `{"title":"New Title","author":"New Author","published_year":2023}`
    req := createRequestWithID("PUT", "/books/1", bytes.NewBufferString(updateBody))
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    h.Update(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)
}

func TestBookHandler_Update_Conflict(t *testing.T) {
    svc := newMockBookService()
    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "nonexistent")
    updateBody := `{"title":"New Title","author":"New Author"}`
    req := createRequestWithID("PUT", "/books/nonexistent", bytes.NewBufferString(updateBody))
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    h.Update(rec, req)
    require.Equal(t, http.StatusConflict, rec.Code)
}

func TestBookHandler_Delete_Success(t *testing.T) {
    svc := newMockBookService()
    svc.books["1"] = &model.Book{ID: "1", Title: "Test Book"}

    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "1")
    req := createRequestWithID("DELETE", "/books/1", nil)
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
    rec := httptest.NewRecorder()

    h.Delete(rec, req)
    require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestBookHandler_Delete_Failure(t *testing.T) {
    svc := newMockBookService()
    h := NewBookHandler(svc)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "nonexistent")
    req := createRequestWithID("DELETE", "/books/nonexistent", nil)
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
    rec := httptest.NewRecorder()

    h.Delete(rec, req)
    require.Equal(t, http.StatusInternalServerError, rec.Code)
}