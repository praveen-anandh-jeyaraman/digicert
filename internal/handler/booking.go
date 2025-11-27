package handler

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
)

type BookingHandler struct {
    bookingSvc service.BookingService
}

func NewBookingHandler(bookingSvc service.BookingService) *BookingHandler {
    return &BookingHandler{bookingSvc: bookingSvc}
}

// Borrow godoc
// @Summary      Borrow a book
// @Description  Borrow a book from the library
// @Tags         Bookings
// @Security     BearerAuth
// @Accept       json
// @Param        request  body      model.BorrowBookRequest  true  "Borrow request"
// @Produce      json
// @Success      201  {object}  model.Booking
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /bookings [post]
func (h *BookingHandler) Borrow(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    userID := GetUserID(r.Context())

    if userID == "" {
        log.Printf("[%s] Unauthorized", requestID)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req model.BorrowBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    errs := ValidationErrors{}
    if req.BookID == "" {
        errs["book_id"] = "book_id is required"
    }
    if req.BorrowDays < 1 || req.BorrowDays > 30 {
        errs["borrow_days"] = "borrow_days must be between 1 and 30"
    }

    if len(errs) > 0 {
        WriteValidationErrors(r.Context(), w, errs)
        return
    }

    booking, err := h.bookingSvc.Borrow(r.Context(), userID, &req)
    if err != nil {
        if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "not found") {
            log.Printf("[%s] Borrow failed: %v", requestID, err)
            WriteError(r.Context(), w, http.StatusConflict, err.Error())
            return
        }
        log.Printf("[%s] Borrow failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to borrow book")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(booking)
    log.Printf("[%s] Book borrowed: %s by user %s", requestID, booking.BookID, userID)
}

// Return godoc
// @Summary      Return a book
// @Description  Return a borrowed book to the library
// @Tags         Bookings
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "Booking ID"
// @Produce      json
// @Success      200  {object}  model.Booking
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /bookings/{id}/return [post]
func (h *BookingHandler) Return(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    userID := GetUserID(r.Context())

    if userID == "" {
        log.Printf("[%s] Unauthorized", requestID)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    bookingID := chi.URLParam(r, "id")
    if bookingID == "" {
        WriteError(r.Context(), w, http.StatusBadRequest, "Booking ID is required")
        return
    }

    booking, err := h.bookingSvc.Return(r.Context(), bookingID)
    if err != nil {
        if strings.Contains(err.Error(), "not found") {
            log.Printf("[%s] Return failed: %v", requestID, err)
            WriteError(r.Context(), w, http.StatusNotFound, "Booking not found")
            return
        }
        log.Printf("[%s] Return failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to return book")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(booking)
    log.Printf("[%s] Book returned: %s by user %s", requestID, booking.BookID, userID)
}

// GetMyBookings godoc
// @Summary      Get my bookings
// @Description  Get all bookings for current user
// @Tags         Bookings
// @Security     BearerAuth
// @Param        limit   query     int     false  "Items per page"  default(20)
// @Param        offset  query     int     false  "Pagination offset"  default(0)
// @Produce      json
// @Success      200  {array}   model.Booking
// @Failure      401  {object}  ErrorResponse
// @Router       /bookings [get]
func (h *BookingHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    userID := GetUserID(r.Context())

    if userID == "" {
        log.Printf("[%s] Unauthorized", requestID)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
        return
    }

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

    bookings, err := h.bookingSvc.GetByUser(r.Context(), userID, limit, offset)
    if err != nil {
        log.Printf("[%s] Get bookings failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to get bookings")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(bookings)
    log.Printf("[%s] Retrieved %d bookings for user %s", requestID, len(bookings), userID)
}

// GetBooking godoc
// @Summary      Get booking details
// @Description  Get details of a specific booking
// @Tags         Bookings
// @Security     BearerAuth
// @Param        id  path  string  true  "Booking ID"
// @Produce      json
// @Success      200  {object}  model.Booking
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /bookings/{id} [get]
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    userID := GetUserID(r.Context())

    if userID == "" {
        log.Printf("[%s] Unauthorized", requestID)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    bookingID := chi.URLParam(r, "id")
    booking, err := h.bookingSvc.GetByID(r.Context(), bookingID)
    if err != nil {
        log.Printf("[%s] Booking not found: %s", requestID, bookingID)
        WriteError(r.Context(), w, http.StatusNotFound, "Booking not found")
        return
    }

    // Users can only see their own bookings
    if booking.UserID != userID {
        log.Printf("[%s] Unauthorized access to booking %s", requestID, bookingID)
        WriteError(r.Context(), w, http.StatusForbidden, "Forbidden")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(booking)
}
// ListAllBookings godoc
// @Summary      List all bookings (admin)
// @Description  Get all bookings in the system
// @Tags         Admin
// @Security     BearerAuth
// @Param        limit   query     int     false  "Items per page"  default(20)
// @Param        offset  query     int     false  "Pagination offset"  default(0)
// @Produce      json
// @Success      200  {array}   model.Booking
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /admin/bookings [get]
func (h *BookingHandler) ListAllBookings(w http.ResponseWriter, r *http.Request) {
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

    bookings, err := h.bookingSvc.List(r.Context(), limit, offset)
    if err != nil {
        log.Printf("[%s] List bookings failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to list bookings")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(bookings)
    log.Printf("[%s] Listed %d bookings", requestID, len(bookings))
}