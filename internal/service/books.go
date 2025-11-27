package service

import (
    "context"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
)

type BookService interface {
    List(ctx context.Context, limit, offset int) ([]model.Book, error)
    GetByID(ctx context.Context, id string) (model.Book, error)
    Create(ctx context.Context, b *model.Book) error
    Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) // ← Changed
    Delete(ctx context.Context, id string) error
}

type bookServiceImpl struct {
    repo repo.BookRepo
}

func NewBookService(r repo.BookRepo) BookService {
    return &bookServiceImpl{repo: r}
}

func (s *bookServiceImpl) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
    return s.repo.List(ctx, limit, offset)
}

func (s *bookServiceImpl) GetByID(ctx context.Context, id string) (model.Book, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *bookServiceImpl) Create(ctx context.Context, b *model.Book) error {
    return s.repo.Create(ctx, b)
}

func (s *bookServiceImpl) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
    return s.repo.Update(ctx, id, updates)
}

func (s *bookServiceImpl) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}