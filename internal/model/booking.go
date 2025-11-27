package model

import "time"

type Booking struct {
    ID         string     `json:"id"`
    UserID     string     `json:"user_id"`
    BookID     string     `json:"book_id"`
    Book       *Book      `json:"book,omitempty"`
    BorrowedAt time.Time  `json:"borrowed_at"`
    DueDate    time.Time  `json:"due_date"`
    ReturnedAt *time.Time `json:"returned_at,omitempty"`
    Status     string     `json:"status"` // ACTIVE, RETURNED, OVERDUE
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
}

type BorrowBookRequest struct {
    BookID     string `json:"book_id" validate:"required"`
    BorrowDays int    `json:"borrow_days" validate:"required,min=1,max=30"`
}

type ReturnBookRequest struct {
    BookingID string `json:"booking_id" validate:"required"`
}

type BorrowBookResponse struct {
    Booking *Booking `json:"booking"`
    Message string   `json:"message"`
}