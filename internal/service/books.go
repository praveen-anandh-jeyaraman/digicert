package service

import (
    "context"
    "errors"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
)

var ErrConflict = errors.New("conflict")

// BookService interface defines the contract for book operations
type BookService interface {
    List(ctx context.Context, limit, offset int) ([]model.Book, error)
    Get(ctx context.Context, id string) (model.Book, error)
    Create(ctx context.Context, b *model.Book) error
    Update(ctx context.Context, b *model.Book) error
    Delete(ctx context.Context, id string) error
}

// bookServiceImpl is the concrete implementation
type bookServiceImpl struct {
    repo repo.BookRepo
}

func NewBookService(r repo.BookRepo) BookService {
    return &bookServiceImpl{repo: r}
}

func (s *bookServiceImpl) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
    return s.repo.List(ctx, limit, offset)
}

func (s *bookServiceImpl) Get(ctx context.Context, id string) (model.Book, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *bookServiceImpl) Create(ctx context.Context, b *model.Book) error {
    return s.repo.Create(ctx, b)
}

func (s *bookServiceImpl) Update(ctx context.Context, b *model.Book) error {
    if err := s.repo.Update(ctx, b); err != nil {
        if err.Error() == "conflict: stale version or not found" {
            return ErrConflict
        }
        return err
    }
    return nil
}

func (s *bookServiceImpl) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}