package repo

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
)

type BookingRepo interface {
    Create(ctx context.Context, b *model.Booking) error
    GetByID(ctx context.Context, id string) (*model.Booking, error)
    GetByUser(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error)
    GetActive(ctx context.Context, userID, bookID string) (*model.Booking, error)
    Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Booking, error)
    MarkOverdue(ctx context.Context) error
    List(ctx context.Context, limit, offset int) ([]model.Booking, error)
}

type pgBookingRepo struct {
    db *pgxpool.Pool
}

func NewBookingRepo(db *pgxpool.Pool) BookingRepo {
    return &pgBookingRepo{db: db}
}

// Create inserts a new booking
func (r *pgBookingRepo) Create(ctx context.Context, b *model.Booking) error {
    if b.ID == "" {
        b.ID = uuid.New().String()
    }
    if b.CreatedAt.IsZero() {
        b.CreatedAt = time.Now().UTC()
    }
    if b.UpdatedAt.IsZero() {
        b.UpdatedAt = time.Now().UTC()
    }

    err := r.db.QueryRow(ctx,
        `INSERT INTO bookings (id, user_id, book_id, borrowed_at, due_date, status, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
         RETURNING id, user_id, book_id, borrowed_at, due_date, returned_at, status, created_at, updated_at`,
        b.ID, b.UserID, b.BookID, b.BorrowedAt, b.DueDate, b.Status, b.CreatedAt, b.UpdatedAt,
    ).Scan(&b.ID, &b.UserID, &b.BookID, &b.BorrowedAt, &b.DueDate, &b.ReturnedAt, &b.Status, &b.CreatedAt, &b.UpdatedAt)

    if err != nil {
        return err
    }
    return nil
}

// GetByID retrieves booking by ID
func (r *pgBookingRepo) GetByID(ctx context.Context, id string) (*model.Booking, error) {
    b := &model.Booking{}
    err := r.db.QueryRow(ctx,
        `SELECT id, user_id, book_id, borrowed_at, due_date, returned_at, status, created_at, updated_at 
         FROM bookings WHERE id = $1`,
        id,
    ).Scan(&b.ID, &b.UserID, &b.BookID, &b.BorrowedAt, &b.DueDate, &b.ReturnedAt, &b.Status, &b.CreatedAt, &b.UpdatedAt)

    if err != nil {
        return nil, errors.New("booking not found")
    }
    return b, nil
}

// GetByUser retrieves user's bookings
func (r *pgBookingRepo) GetByUser(ctx context.Context, userID string, limit, offset int) ([]model.Booking, error) {
    rows, err := r.db.Query(ctx,
        `SELECT id, user_id, book_id, borrowed_at, due_date, returned_at, status, created_at, updated_at 
         FROM bookings WHERE user_id = $1 
         ORDER BY borrowed_at DESC LIMIT $2 OFFSET $3`,
        userID, limit, offset,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var bookings []model.Booking
    for rows.Next() {
        b := model.Booking{}
        if err := rows.Scan(&b.ID, &b.UserID, &b.BookID, &b.BorrowedAt, &b.DueDate, &b.ReturnedAt, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
            return nil, err
        }
        bookings = append(bookings, b)
    }
    return bookings, nil
}

// GetActive retrieves active booking for user+book
func (r *pgBookingRepo) GetActive(ctx context.Context, userID, bookID string) (*model.Booking, error) {
    b := &model.Booking{}
    err := r.db.QueryRow(ctx,
        `SELECT id, user_id, book_id, borrowed_at, due_date, returned_at, status, created_at, updated_at 
         FROM bookings WHERE user_id = $1 AND book_id = $2 AND status = 'ACTIVE'`,
        userID, bookID,
    ).Scan(&b.ID, &b.UserID, &b.BookID, &b.BorrowedAt, &b.DueDate, &b.ReturnedAt, &b.Status, &b.CreatedAt, &b.UpdatedAt)

    if err != nil {
        return nil, errors.New("no active booking found")
    }
    return b, nil
}

// Update updates booking
func (r *pgBookingRepo) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Booking, error) {
    updates["updated_at"] = time.Now().UTC()

    // Build dynamic query
    query := `UPDATE bookings SET `
    args := []interface{}{}
    i := 1

    for key, value := range updates {
        if i > 1 {
            query += ", "
        }
        query += key + "=$" + string(rune(i+48))
        args = append(args, value)
        i++
    }

    query += ` WHERE id = $` + string(rune(i+48))
    args = append(args, id)
    query += ` RETURNING id, user_id, book_id, borrowed_at, due_date, returned_at, status, created_at, updated_at`

    b := &model.Booking{}
    err := r.db.QueryRow(ctx, query, args...).Scan(&b.ID, &b.UserID, &b.BookID, &b.BorrowedAt, &b.DueDate, &b.ReturnedAt, &b.Status, &b.CreatedAt, &b.UpdatedAt)
    if err != nil {
        return nil, err
    }

    return b, nil
}

// MarkOverdue marks overdue bookings
func (r *pgBookingRepo) MarkOverdue(ctx context.Context) error {
    _, err := r.db.Exec(ctx,
        `UPDATE bookings SET status = 'OVERDUE', updated_at = NOW() 
         WHERE status = 'ACTIVE' AND due_date < NOW()`,
    )
    return err
}

// List retrieves all bookings (admin)
func (r *pgBookingRepo) List(ctx context.Context, limit, offset int) ([]model.Booking, error) {
    rows, err := r.db.Query(ctx,
        `SELECT id, user_id, book_id, borrowed_at, due_date, returned_at, status, created_at, updated_at 
         FROM bookings ORDER BY borrowed_at DESC LIMIT $1 OFFSET $2`,
        limit, offset,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var bookings []model.Booking
    for rows.Next() {
        b := model.Booking{}
        if err := rows.Scan(&b.ID, &b.UserID, &b.BookID, &b.BorrowedAt, &b.DueDate, &b.ReturnedAt, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
            return nil, err
        }
        bookings = append(bookings, b)
    }
    return bookings, nil
}