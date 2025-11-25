package handler

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
)

type BookHandler struct {
    svc    service.BookService
    logger *StructuredLogger
}

func NewBookHandler(svc service.BookService) *BookHandler {
    return &BookHandler{
        svc:    svc,
        logger: NewStructuredLogger(),
    }
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
    ctx := r.Context()
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Listing books", requestID)

    limit := 20
    offset := 0
    qs := r.URL.Query()

    if v := qs.Get("limit"); v != "" {
        i, err := strconv.Atoi(v)
        if err != nil || i < 1 || i > 100 {
            WriteError(ctx, w, http.StatusBadRequest, "limit must be between 1 and 100")
            return
        }
        limit = i
    }

    if v := qs.Get("offset"); v != "" {
        i, err := strconv.Atoi(v)
        if err != nil || i < 0 {
            WriteError(ctx, w, http.StatusBadRequest, "offset must be >= 0")
            return
        }
        offset = i
    }

    books, err := h.svc.List(ctx, limit, offset)
    if err != nil {
        log.Printf("[%s] Error listing books: %v", requestID, err)
        WriteError(ctx, w, http.StatusInternalServerError, "failed to list books")
        return
    }

    log.Printf("[%s] Successfully listed %d books", requestID, len(books))
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(books); err != nil {
        log.Printf("[%s] failed to encode response: %v", GetRequestID(r.Context()), err)
    }
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
    ctx := r.Context()
    id := chi.URLParam(r, "id")
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Getting book: %s", requestID, id)

    b, err := h.svc.Get(ctx, id)
    if err != nil {
        log.Printf("[%s] Book not found: %s", requestID, id)
        WriteError(ctx, w, http.StatusNotFound, "book not found")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(b); err != nil {
        log.Printf("[%s] failed to encode response: %v", GetRequestID(r.Context()), err)
    }
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
    ctx := r.Context()
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Creating book", requestID)

    var req model.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Failed to decode request: %v", requestID, err)
        WriteError(ctx, w, http.StatusBadRequest, "invalid JSON payload")
        return
    }

    errs := ValidationErrors{}
    req.Title = trim(req.Title)
    req.Author = trim(req.Author)
    req.ISBN = trim(req.ISBN)

    if req.Title == "" {
        errs["title"] = "title is required"
    } else if len(req.Title) > 255 {
        errs["title"] = "title must be at most 255 characters"
    }

    if req.Author == "" {
        errs["author"] = "author is required"
    } else if len(req.Author) > 100 {
        errs["author"] = "author must be at most 100 characters"
    }

    curYear := time.Now().Year()
    if req.PublishedYear != 0 {
        if req.PublishedYear < 1000 || req.PublishedYear > curYear+1 {
            errs["published_year"] = "invalid published_year"
        }
    }

    if len(errs) > 0 {
        log.Printf("[%s] Validation failed: %v", requestID, errs)
        WriteValidationErrors(ctx, w, errs)
        return
    }

    b := model.Book{
        Title:         req.Title,
        Author:        req.Author,
        ISBN:          req.ISBN,
        PublishedYear: req.PublishedYear,
    }

    if err := h.svc.Create(ctx, &b); err != nil {
        log.Printf("[%s] Failed to create book: %v", requestID, err)
        WriteError(ctx, w, http.StatusInternalServerError, "failed to create book")
        return
    }

    log.Printf("[%s] Book created successfully: %s", requestID, b.ID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(b); err != nil {
        log.Printf("[%s] failed to encode response: %v", GetRequestID(r.Context()), err)
    }
}
// Update godoc
// @Summary      Update a book
// @Description  Update book details by ID
// @Tags         Books
// @Accept       json
// @Param        id       path      string  true  "Book ID"
// @Param        request  body      model.Book  true  "Updated book data"
// @Produce      json
// @Success      200  {object}  model.Book
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /books/{id} [put]
func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := chi.URLParam(r, "id")
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Updating book: %s", requestID, id)

    var b model.Book
    if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
        log.Printf("[%s] Failed to decode request: %v", requestID, err)
        WriteError(ctx, w, http.StatusBadRequest, "invalid JSON payload")
        return
    }

    curYear := time.Now().Year()
    if b.PublishedYear != 0 {
        if b.PublishedYear < 1000 || b.PublishedYear > curYear+1 {
            WriteError(ctx, w, http.StatusBadRequest, "invalid published_year")
            return
        }
    }

    b.ID = id
    if err := h.svc.Update(ctx, &b); err != nil {
        if err == service.ErrConflict {
            log.Printf("[%s] Conflict updating book: %s", requestID, id)
            WriteError(ctx, w, http.StatusConflict, "version conflict or book not found")
            return
        }
        log.Printf("[%s] Failed to update book: %v", requestID, err)
        WriteError(ctx, w, http.StatusInternalServerError, "failed to update book")
        return
    }

    log.Printf("[%s] Book updated successfully: %s", requestID, b.ID)
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(b)
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
    ctx := r.Context()
    id := chi.URLParam(r, "id")
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Deleting book: %s", requestID, id)

    if err := h.svc.Delete(ctx, id); err != nil {
        log.Printf("[%s] Failed to delete book: %v", requestID, err)
        WriteError(ctx, w, http.StatusInternalServerError, "failed to delete book")
        return
    }

    log.Printf("[%s] Book deleted successfully: %s", requestID, id)
    w.WriteHeader(http.StatusNoContent)
}

// Health godoc
// @Summary      Health check
// @Description  Check if the service is healthy
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /healthz [get]
func (h *BookHandler) Health(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Health check", requestID)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}