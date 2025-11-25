package test

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/handler"
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
    ctx := context.WithValue(req.Context(), handler.RequestIDKey, "integration-test-123")
    return req.WithContext(ctx)
}

// mockBookService for integration tests
type mockBookService struct {
    books   map[string]*model.Book
    idCount int
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
    m.idCount++
    b.ID = fmt.Sprintf("book-%d", m.idCount)
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
    return &mockBookService{books: make(map[string]*model.Book), idCount: 0}
}

// Integration Tests

func TestIntegration_CreateAndRetrieveBook(t *testing.T) {
    svc := newMockBookService()
    h := handler.NewBookHandler(svc)

    // Create a book
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

    // Retrieve the book
    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", created.ID)
    getReq := createRequestWithID("GET", "/books/"+created.ID, nil)
    getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, chiCtx))
    getRec := httptest.NewRecorder()

    h.Get(getRec, getReq)
    require.Equal(t, http.StatusOK, getRec.Code)

    var retrieved model.Book
    require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &retrieved))
    require.Equal(t, created.ID, retrieved.ID)
    require.Equal(t, "Go Programming", retrieved.Title)
}

func TestIntegration_CreateUpdateDelete(t *testing.T) {
    svc := newMockBookService()
    h := handler.NewBookHandler(svc)

    // Create
    createBody := `{"title":"Rust Book","author":"Jane Smith"}`
    req := createRequestWithID("POST", "/books", bytes.NewBufferString(createBody))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    h.Create(rec, req)
    require.Equal(t, http.StatusCreated, rec.Code)

    var created model.Book
    err := json.Unmarshal(rec.Body.Bytes(), &created)
    require.NoError(t, err)
    // Update
    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", created.ID)
    updateBody := `{"title":"Rust Book 2nd Edition","author":"Jane Smith","published_year":2023}`
    updateReq := createRequestWithID("PUT", "/books/"+created.ID, bytes.NewBufferString(updateBody))
    updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), chi.RouteCtxKey, chiCtx))
    updateRec := httptest.NewRecorder()
    h.Update(updateRec, updateReq)
    require.Equal(t, http.StatusOK, updateRec.Code)

    // Delete
    chiCtx2 := chi.NewRouteContext()
    chiCtx2.URLParams.Add("id", created.ID)
    delReq := createRequestWithID("DELETE", "/books/"+created.ID, nil)
    delReq = delReq.WithContext(context.WithValue(delReq.Context(), chi.RouteCtxKey, chiCtx2))
    delRec := httptest.NewRecorder()
    h.Delete(delRec, delReq)
    require.Equal(t, http.StatusNoContent, delRec.Code)
}

func TestIntegration_ListBooks(t *testing.T) {
    svc := newMockBookService()
    h := handler.NewBookHandler(svc)

    // Create multiple books
    for i := 1; i <= 3; i++ {
        title := "Book " + string(rune(48+i))
        createBody := `{"title":"` + title + `","author":"Author"}`
        req := createRequestWithID("POST", "/books", bytes.NewBufferString(createBody))
        req.Header.Set("Content-Type", "application/json")
        rec := httptest.NewRecorder()
        h.Create(rec, req)
    }

    // List books
    listReq := createRequestWithID("GET", "/books?limit=10&offset=0", nil)
    listRec := httptest.NewRecorder()
    h.List(listRec, listReq)
    require.Equal(t, http.StatusOK, listRec.Code)

    var books []model.Book
    require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &books))
    require.Len(t, books, 3)
}