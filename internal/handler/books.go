package handler

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/logger"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
)

type BookHandler struct {
    svc service.BookService
}

func NewBookHandler(svc service.BookService) *BookHandler {
    return &BookHandler{svc: svc}
}

// UpdateBookRequest for PUT requests
type UpdateBookRequest struct {
    Title         string `json:"title"`
    Author        string `json:"author"`
    PublishedYear int    `json:"published_year"`
    ISBN          string `json:"isbn"`
}

// List godoc
// @Summary      List all books
// @Description  Get a paginated list of all books
// @Tags         Books
// @Param        limit   query     int     false  "Items per page (1-100)"  default(20)
// @Param        offset  query     int     false  "Pagination offset"       default(0)
// @Produce      json
// @Success      200  {array}   model.Book
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /books [get]
func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    limit := 20
    offset := 0

    if l := r.URL.Query().Get("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
            limit = parsed
        }
    }

    if o := r.URL.Query().Get("offset"); o != "" {
        if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
            offset = parsed
        }
    }

    books, err := h.svc.List(r.Context(), limit, offset)
    if err != nil {
        log.Printf("[%s] List failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to list books")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(books)
    log.Printf("[%s] Listed %d books", requestID, len(books))
}

// Get godoc
// @Summary      Get a book by ID
// @Description  Retrieve a single book by its ID
// @Tags         Books
// @Param        id   path      string  true  "Book ID"
// @Produce      json
// @Success      200  {object}  model.Book
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /books/{id} [get]
func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    book, err := h.svc.GetByID(r.Context(), id) // ← Changed from Get to GetByID
    if err != nil {
        if strings.Contains(err.Error(), "not found") {
            log.Printf("[%s] Book not found: %s", requestID, id)
            WriteError(r.Context(), w, http.StatusNotFound, "Book not found")
            return
        }
        log.Printf("[%s] Get failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to get book")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(book)
    log.Printf("[%s] Book retrieved: %s", requestID, id)
}

// Create godoc
// @Summary      Create a new book
// @Description  Create a new book with validation
// @Tags         Books
// @Accept       json
// @Param        request  body      model.CreateBookRequest  true  "Book request"
// @Produce      json
// @Success      201  {object}  model.Book
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /books [post]
func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    var req model.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }
            req.Title = trim(req.Title)
            req.Author = trim(req.Author)
            req.ISBN = trim(req.ISBN)
    book := &model.Book{
        Title:         req.Title,
        Author:        req.Author,
        PublishedYear: req.PublishedYear,
        ISBN:          req.ISBN,
    }

    if err := h.svc.Create(r.Context(), book); err != nil {
        log.Printf("[%s] Create failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to create book")
        return
    }

    // track metric in CloudWatch
      cwLogger := logger.GetLogger()
    if cwLogger != nil {
        _ = cwLogger.PutMetric(r.Context(), "BookCreated", 1, "Count")
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(book)
    log.Printf("[%s] Book created: %s", requestID, book.ID)
}

// Update godoc
// @Summary      Update a book
// @Description  Update book details by ID
// @Tags         Books
// @Accept       json
// @Param        id       path      string  true  "Book ID"
// @Param        request  body      UpdateBookRequest  true  "Updated book data"
// @Produce      json
// @Success      200  {object}  model.Book
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /books/{id} [put]
func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    var req UpdateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    updates := map[string]interface{}{
        "title":          req.Title,
        "author":         req.Author,
        "published_year": req.PublishedYear,
        "isbn":           req.ISBN,
    }

    book, err := h.svc.Update(r.Context(), id, updates)
    if err != nil {
        if strings.Contains(err.Error(), "conflict") {
            log.Printf("[%s] Conflict: %v", requestID, err)
            WriteError(r.Context(), w, http.StatusConflict, "Book was modified by another request. Please refetch and retry.")
            return
        }
        log.Printf("[%s] Update failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to update book")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(book)
    log.Printf("[%s] Book updated: %s", requestID, id)
}

// Delete godoc
// @Summary      Delete a book
// @Description  Delete a book by ID
// @Tags         Books
// @Param        id   path  string  true  "Book ID"
// @Success      204
// @Failure      500  {object}  ErrorResponse
// @Router       /books/{id} [delete]
func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    if err := h.svc.Delete(r.Context(), id); err != nil {
        log.Printf("[%s] Delete failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to delete book")
        return
    }

    w.WriteHeader(http.StatusNoContent)
    log.Printf("[%s] Book deleted: %s", requestID, id)
}