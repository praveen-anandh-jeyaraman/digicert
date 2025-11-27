package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/stretchr/testify/require"
)

// Mock booking service
type mockBookingService struct {
    borrowFn    func(ctx context.Context, userID string, req *model.BorrowBookRequest) (*model.Booking, error)
    returnFn    func(ctx context.Context, bookingID string) (*model.Booking, error)
    getByUserFn func(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error)
    getByIDFn   func(ctx context.Context, id string) (*model.Booking, error)
    listFn      func(ctx context.Context, limit, offset int) ([]model.Booking, error)
    updateFn    func(ctx context.Context) error
}

func (m *mockBookingService) Borrow(ctx context.Context, userID string, req *model.BorrowBookRequest) (*model.Booking, error) {
    return m.borrowFn(ctx, userID, req)
}

func (m *mockBookingService) Return(ctx context.Context, bookingID string) (*model.Booking, error) {
    return m.returnFn(ctx, bookingID)
}

func (m *mockBookingService) GetByUser(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error) {
    return m.getByUserFn(ctx, userID, limit, offset)
}

func (m *mockBookingService) GetByID(ctx context.Context, id string) (*model.Booking, error) {
    return m.getByIDFn(ctx, id)
}

func (m *mockBookingService) List(ctx context.Context, limit, offset int) ([]model.Booking, error) {
    return m.listFn(ctx, limit, offset)
}

func (m *mockBookingService) UpdateOverdue(ctx context.Context) error {
    return m.updateFn(ctx)
}

// Helper to set request ID in context properly
func createBookingTestRequest(method, path string, body string, requestID string) *http.Request {
    req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    // Use RequestIDKey (typed context key)
    ctx := context.WithValue(req.Context(), RequestIDKey, requestID)
    return req.WithContext(ctx)
}

func TestBookingHandler_Borrow_Success(t *testing.T) {
    now := time.Now().UTC()
    mock := &mockBookingService{
        borrowFn: func(_ context.Context, userID string, req *model.BorrowBookRequest) (*model.Booking, error) {
            return &model.Booking{
                ID:         "booking-1",
                UserID:     userID,
                BookID:     req.BookID,
                BorrowedAt: now,
                DueDate:    now.AddDate(0, 0, req.BorrowDays),
                Status:     "ACTIVE",
                CreatedAt:  now,
                UpdatedAt:  now,
            }, nil
        },
    }
    h := NewBookingHandler(mock)

    req := createBookingTestRequest("POST", "/bookings", `{"book_id":"book-1","borrow_days":14}`, "test-booking-borrow-001")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "user_id", "user-1")
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Borrow(rec, req)
    require.Equal(t, http.StatusCreated, rec.Code)

    var booking model.Booking
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &booking))
    require.Equal(t, "ACTIVE", booking.Status)
    require.Equal(t, "user-1", booking.UserID)
}

func TestBookingHandler_Borrow_InvalidDays(t *testing.T) {
    mock := &mockBookingService{}
    h := NewBookingHandler(mock)

    req := createBookingTestRequest("POST", "/bookings", `{"book_id":"book-1","borrow_days":60}`, "test-booking-borrow-002")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "user_id", "user-1")
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Borrow(rec, req)
    require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBookingHandler_Return_Success(t *testing.T) {
    now := time.Now().UTC()
    mock := &mockBookingService{
        returnFn: func(_ context.Context, bookingID string) (*model.Booking, error) {
            return &model.Booking{
                ID:         bookingID,
                UserID:     "user-1",
                BookID:     "book-1",
                BorrowedAt: now.AddDate(0, 0, -14),
                DueDate:    now,
                ReturnedAt: &now,
                Status:     "RETURNED",
                CreatedAt:  now.AddDate(0, 0, -14),
                UpdatedAt:  now,
            }, nil
        },
    }
    h := NewBookingHandler(mock)

    chiCtx := chi.NewRouteContext()
    chiCtx.URLParams.Add("id", "booking-1")
    req := createBookingTestRequest("POST", "/bookings/booking-1/return", "", "test-booking-return-001")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "user_id", "user-1")
    ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.Return(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var booking model.Booking
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &booking))
    require.Equal(t, "RETURNED", booking.Status)
}

func TestBookingHandler_GetMyBookings_Success(t *testing.T) {
    mock := &mockBookingService{
        getByUserFn: func(_ context.Context, userID string, limit, offset int) ([]model.Booking, error) {
            return []model.Booking{
                {
                    ID:     "booking-1",
                    UserID: userID,
                    BookID: "book-1",
                    Status: "ACTIVE",
                },
            }, nil
        },
    }
    h := NewBookingHandler(mock)

    req := createBookingTestRequest("GET", "/bookings", "", "test-booking-getmy-001")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "user_id", "user-1")
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.GetMyBookings(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var bookings []model.Booking
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &bookings))
    require.Len(t, bookings, 1)
}

func TestBookingHandler_ListAllBookings_Success(t *testing.T) {
    mock := &mockBookingService{
        listFn: func(_ context.Context, limit, offset int) ([]model.Booking, error) {
            return []model.Booking{
                {ID: "1", UserID: "user-1", Status: "ACTIVE"},
                {ID: "2", UserID: "user-2", Status: "RETURNED"},
            }, nil
        },
    }
    h := NewBookingHandler(mock)

    req := createBookingTestRequest("GET", "/admin/bookings", "", "test-booking-listall-001")
    ctx := req.Context()
    ctx = context.WithValue(ctx, "role", "ADMIN")
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.ListAllBookings(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var bookings []model.Booking
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &bookings))
    require.Len(t, bookings, 2)
}