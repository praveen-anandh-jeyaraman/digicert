package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/praveen-anandh-jeyaraman/digicert/internal/model"
)

type BookRepo interface {
	List(ctx context.Context, limit, offset int) ([]model.Book, error)
	GetByID(ctx context.Context, id string) (model.Book, error)
	Create(ctx context.Context, b *model.Book) error
    Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) // ← Changed
	Delete(ctx context.Context, id string) error
}

type pgBookRepo struct {
	db *pgxpool.Pool
}

func NewBookRepo(db *pgxpool.Pool) BookRepo {
	return &pgBookRepo{db: db}
}

func (r *pgBookRepo) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
	rows, err := r.db.Query(ctx, `SELECT id,title,author,published_year,isbn,created_at,updated_at,version FROM books ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Book
	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.PublishedYear, &b.ISBN, &b.CreatedAt, &b.UpdatedAt, &b.Version); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *pgBookRepo) GetByID(ctx context.Context, id string) (model.Book, error) {
	var b model.Book
	err := r.db.QueryRow(ctx, `SELECT id,title,author,published_year,isbn,created_at,updated_at,version FROM books WHERE id=$1`, id).Scan(
		&b.ID, &b.Title, &b.Author, &b.PublishedYear, &b.ISBN, &b.CreatedAt, &b.UpdatedAt, &b.Version)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (r *pgBookRepo) Create(ctx context.Context, b *model.Book) error {
	now := time.Now().UTC()
	err := r.db.QueryRow(ctx,
		`INSERT INTO books (title,author,published_year,isbn,created_at,updated_at,version) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at,updated_at,version`,
		b.Title, b.Author, b.PublishedYear, b.ISBN, now, now, 1).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt, &b.Version)
	return err
}

func (r *pgBookRepo) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
    // Step 1: Get current book (including version)
    var currentBook model.Book
    err := r.db.QueryRow(ctx,
        `SELECT id, version FROM books WHERE id = $1`,
        id,
    ).Scan(&currentBook.ID, &currentBook.Version)
    if err != nil {
        return nil, errors.New("book not found")
    }

    // Step 2: Increment version
    newVersion := currentBook.Version + 1

    // Step 3: Update with optimistic locking
    cmdTag, err := r.db.Exec(ctx,
        `UPDATE books 
         SET title=$1, author=$2, published_year=$3, isbn=$4, 
             updated_at=$5, version=$6
         WHERE id=$7 AND version=$8`,
        updates["title"], updates["author"], updates["published_year"], updates["isbn"],
        time.Now().UTC(), newVersion, id, currentBook.Version,
    )

    if cmdTag.RowsAffected() == 0 {
        return nil, errors.New("conflict: book was modified by another request")
    }

    // Return updated book
    book, err := r.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return &book, nil
}

func (r *pgBookRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM books WHERE id=$1`, id)
	return err
}
