package repo

import (
    "context"
    "errors"
    "time"
	"fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
)

type UserRepo interface {
    Create(ctx context.Context, u *model.User) error
    GetByID(ctx context.Context, id string) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]model.User, error)
}

type pgUserRepo struct {
    db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepo {
    return &pgUserRepo{db: db}
}

// Create inserts a new user
func (r *pgUserRepo) Create(ctx context.Context, u *model.User) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    if u.CreatedAt.IsZero() {
        u.CreatedAt = time.Now().UTC()
    }
    if u.UpdatedAt.IsZero() {
        u.UpdatedAt = time.Now().UTC()
    }

    err := r.db.QueryRow(ctx,
        `INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, username, email, role, created_at, updated_at`,
        u.ID, u.Username, u.Email, u.Password, u.Role, u.CreatedAt, u.UpdatedAt,
    ).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)

    if err != nil {
        if err.Error() == "duplicate key value violates unique constraint \"users_username_key\"" {
            return errors.New("username already exists")
        }
        if err.Error() == "duplicate key value violates unique constraint \"users_email_key\"" {
            return errors.New("email already exists")
        }
        return err
    }

    return nil
}

// In GetByID method
func (r *pgUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
    u := &model.User{}
    err := r.db.QueryRow(ctx,
        `SELECT id, username, email, role, created_at, updated_at FROM users WHERE id = $1`,
        id,
    ).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)

    if err != nil {
        return nil, errors.New("user not found")
    }
    return u, nil
}

// In GetByUsername method (for login)
func (r *pgUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    u := &model.User{}
    err := r.db.QueryRow(ctx,
        `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE username = $1`,
        username,
    ).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)

    if err != nil {
        return nil, errors.New("user not found")
    }
    return u, nil
}

// GetByEmail retrieves user by email
func (r *pgUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    u := &model.User{}
    err := r.db.QueryRow(ctx,
        `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`,
        email,
    ).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)

    if err != nil {
        return nil, errors.New("user not found")
    }
    return u, nil
}

// Update updates user information
func (r *pgUserRepo) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
    u := &model.User{}
    updates["updated_at"] = time.Now().UTC()

    // Build dynamic query
    query := `UPDATE users SET `
    args := []interface{}{}
    i := 1

    for key, value := range updates {
        if i > 1 {
            query += ", "
        }
        query += key + " = $" + fmt.Sprintf("%d", i)
        args = append(args, value)
        i++
    }

    query += ` WHERE id = $` + fmt.Sprintf("%d", i)
    args = append(args, id)

    query += ` RETURNING id, username, email, created_at, updated_at`

    err := r.db.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
    if err != nil {
        return nil, err
    }

    return u, nil
}

// Delete removes a user
func (r *pgUserRepo) Delete(ctx context.Context, id string) error {
    cmdTag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
    if err != nil {
        return err
    }
    if cmdTag.RowsAffected() == 0 {
        return errors.New("user not found")
    }
    return nil
}

// List retrieves all users (paginated)
func (r *pgUserRepo) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    rows, err := r.db.Query(ctx,
        `SELECT id, username, email,role, created_at, updated_at FROM users 
         ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
        limit, offset,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []model.User
    for rows.Next() {
        u := model.User{}
        if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    return users, nil
}