package service

import (
    "context"
    "errors"
    "time"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
)

type BookingService interface {
    Borrow(ctx context.Context, userID string, req *model.BorrowBookRequest) (*model.Booking, error)
    Return(ctx context.Context, bookingID string) (*model.Booking, error)
    GetByUser(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error)
    GetByID(ctx context.Context, id string) (*model.Booking, error)
    List(ctx context.Context, limit, offset int) ([]model.Booking, error)
    UpdateOverdue(ctx context.Context) error
}

type bookingService struct {
    bookingRepo repo.BookingRepo
    bookRepo    repo.BookRepo
    userRepo    repo.UserRepo
}

func NewBookingService(br repo.BookingRepo, bk repo.BookRepo, u repo.UserRepo) BookingService {
    return &bookingService{
        bookingRepo: br,
        bookRepo:    bk,
        userRepo:    u,
    }
}

func (s *bookingService) Borrow(ctx context.Context, userID string, req *model.BorrowBookRequest) (*model.Booking, error) {
    _, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, errors.New("user not found")
    }

    _, err = s.bookRepo.GetByID(ctx, req.BookID)
    if err != nil {
        return nil, errors.New("book not found")
    }

    active, _ := s.bookingRepo.GetActive(ctx, userID, req.BookID)
    if active != nil {
        return nil, errors.New("you already have an active booking for this book")
    }

    if req.BorrowDays < 1 || req.BorrowDays > 30 {
        return nil, errors.New("borrow days must be between 1 and 30")
    }

    booking := &model.Booking{
        UserID:     userID,
        BookID:     req.BookID,
        BorrowedAt: time.Now().UTC(),
        DueDate:    time.Now().UTC().AddDate(0, 0, req.BorrowDays),
        Status:     "ACTIVE",
    }

    if err := s.bookingRepo.Create(ctx, booking); err != nil {
        return nil, err
    }

    return booking, nil
}

func (s *bookingService) Return(ctx context.Context, bookingID string) (*model.Booking, error) {
    booking, err := s.bookingRepo.GetByID(ctx, bookingID)
    if err != nil {
        return nil, errors.New("booking not found")
    }

    if booking.Status == "RETURNED" {
        return nil, errors.New("book already returned")
    }

    now := time.Now().UTC()
    updates := map[string]interface{}{
        "returned_at": now,
        "status":      "RETURNED",
    }

    return s.bookingRepo.Update(ctx, bookingID, updates)
}

// GetByUser retrieves user's bookings
func (s *bookingService) GetByUser(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error) {
    return s.bookingRepo.GetByUser(ctx, userID, limit, offset)
}

// GetByID retrieves booking by ID
func (s *bookingService) GetByID(ctx context.Context, id string) (*model.Booking, error) {
    return s.bookingRepo.GetByID(ctx, id)
}

// List retrieves all bookings
func (s *bookingService) List(ctx context.Context, limit, offset int) ([]model.Booking, error) {
    return s.bookingRepo.List(ctx, limit, offset)
}

// UpdateOverdue marks overdue bookings
func (s *bookingService) UpdateOverdue(ctx context.Context) error {
    return s.bookingRepo.MarkOverdue(ctx)
}