package service

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
    "github.com/stretchr/testify/require"
)

// Mock repos
type mockBookingRepoForTest struct {
    createFn    func(ctx context.Context, b *model.Booking) error
    getByIDFn   func(ctx context.Context, id string) (*model.Booking, error)
    getByUserFn func(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error)
    getActiveFn func(ctx context.Context, userID, bookID string) (*model.Booking, error)
    updateFn    func(ctx context.Context, id string, updates map[string]interface{}) (*model.Booking, error)
    listFn      func(ctx context.Context, limit, offset int) ([]model.Booking, error)
    markOverdueFn func(ctx context.Context) error
}

func (m *mockBookingRepoForTest) Create(ctx context.Context, b *model.Booking) error {
    return m.createFn(ctx, b)
}
func (m *mockBookingRepoForTest) GetByID(ctx context.Context, id string) (*model.Booking, error) {
    return m.getByIDFn(ctx, id)
}
func (m *mockBookingRepoForTest) GetByUser(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error) {
    return m.getByUserFn(ctx, userID, limit, offset)
}
func (m *mockBookingRepoForTest) GetActive(ctx context.Context, userID, bookID string) (*model.Booking, error) {
    return m.getActiveFn(ctx, userID, bookID)
}
func (m *mockBookingRepoForTest) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Booking, error) {
    return m.updateFn(ctx, id, updates)
}
func (m *mockBookingRepoForTest) List(ctx context.Context, limit, offset int) ([]model.Booking, error) {
    return m.listFn(ctx, limit, offset)
}
func (m *mockBookingRepoForTest) MarkOverdue(ctx context.Context) error {
    return m.markOverdueFn(ctx)
}

var _ repo.BookingRepo = (*mockBookingRepoForTest)(nil)

type mockBookRepoForTest struct {
    getByIDFn func(ctx context.Context, id string) (model.Book, error)
    createFn  func(ctx context.Context, b *model.Book) error
    listFn    func(ctx context.Context, limit, offset int) ([]model.Book, error)
    updateFn  func(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error)
    deleteFn  func(ctx context.Context, id string) error
}

func (m *mockBookRepoForTest) GetByID(ctx context.Context, id string) (model.Book, error) {
    return m.getByIDFn(ctx, id)
}
func (m *mockBookRepoForTest) Create(ctx context.Context, b *model.Book) error {
    return m.createFn(ctx, b)
}
func (m *mockBookRepoForTest) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
    return m.listFn(ctx, limit, offset)
}
func (m *mockBookRepoForTest) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
    return m.updateFn(ctx, id, updates)
}
func (m *mockBookRepoForTest) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

var _ repo.BookRepo = (*mockBookRepoForTest)(nil)

type mockUserRepoForTest struct {
    getByIDFn       func(ctx context.Context, id string) (*model.User, error)
    getByUsernameFn func(ctx context.Context, username string) (*model.User, error)
    getByEmailFn    func(ctx context.Context, email string) (*model.User, error)
    createFn        func(ctx context.Context, u *model.User) error
    updateFn        func(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
    listFn          func(ctx context.Context, limit, offset int) ([]model.User, error)
    deleteFn        func(ctx context.Context, id string) error
}

func (m *mockUserRepoForTest) GetByID(ctx context.Context, id string) (*model.User, error) {
    return m.getByIDFn(ctx, id)
}
func (m *mockUserRepoForTest) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return m.getByUsernameFn(ctx, username)
}
func (m *mockUserRepoForTest) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    return m.getByEmailFn(ctx, email)
}
func (m *mockUserRepoForTest) Create(ctx context.Context, u *model.User) error {
    return m.createFn(ctx, u)
}
func (m *mockUserRepoForTest) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
    return m.updateFn(ctx, id, updates)
}
func (m *mockUserRepoForTest) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    return m.listFn(ctx, limit, offset)
}
func (m *mockUserRepoForTest) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

var _ repo.UserRepo = (*mockUserRepoForTest)(nil)

func TestBookingService_Borrow_Success(t *testing.T) {
    ctx := context.Background()
    now := time.Now().UTC()

    bookingRepo := &mockBookingRepoForTest{
        getActiveFn: func(_ context.Context, userID, bookID string) (*model.Booking, error) {
            return nil, errors.New("no active booking")
        },
        createFn: func(_ context.Context, b *model.Booking) error {
            b.ID = "booking-1"
            b.CreatedAt = now
            b.UpdatedAt = now
            b.Status = "ACTIVE"
            return nil
        },
    }

    userRepo := &mockUserRepoForTest{
        getByIDFn: func(_ context.Context, id string) (*model.User, error) {
            return &model.User{ID: id, Username: "john"}, nil
        },
    }

    bookRepo := &mockBookRepoForTest{
        getByIDFn: func(_ context.Context, id string) (model.Book, error) {
            return model.Book{ID: id, Title: "Go Programming"}, nil
        },
    }

    svc := NewBookingService(bookingRepo, bookRepo, userRepo)
    req := &model.BorrowBookRequest{BookID: "book-1", BorrowDays: 14}
    booking, err := svc.Borrow(ctx, "user-1", req)

    require.NoError(t, err)
    require.Equal(t, "ACTIVE", booking.Status)
    require.NotEmpty(t, booking.ID)
}

func TestBookingService_Return_Success(t *testing.T) {
    ctx := context.Background()
    now := time.Now().UTC()

    bookingRepo := &mockBookingRepoForTest{
        getByIDFn: func(_ context.Context, id string) (*model.Booking, error) {
            return &model.Booking{
                ID:     id,
                Status: "ACTIVE",
            }, nil
        },
        updateFn: func(_ context.Context, id string, updates map[string]interface{}) (*model.Booking, error) {
            return &model.Booking{
                ID:         id,
                Status:     "RETURNED",
                ReturnedAt: &now,
            }, nil
        },
    }

    svc := NewBookingService(bookingRepo, nil, nil)
    booking, err := svc.Return(ctx, "booking-1")

    require.NoError(t, err)
    require.Equal(t, "RETURNED", booking.Status)
    require.NotNil(t, booking.ReturnedAt)
}

func TestBookingService_GetByUser_Success(t *testing.T) {
    ctx := context.Background()

    bookingRepo := &mockBookingRepoForTest{
        getByUserFn: func(_ context.Context, userID string, limit, offset int) ([]model.Booking, error) {
            return []model.Booking{
                {ID: "1", UserID: userID, Status: "ACTIVE"},
            }, nil
        },
    }

    svc := NewBookingService(bookingRepo, nil, nil)
    bookings, err := svc.GetByUser(ctx, "user-1", 10, 0)

    require.NoError(t, err)
    require.Len(t, bookings, 1)
}