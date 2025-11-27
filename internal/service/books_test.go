package service

import (
    "context"
    "errors"
    "testing"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
    "github.com/stretchr/testify/require"
)

// Mock book repo
type mockBookRepo struct {
    createFn   func(ctx context.Context, b *model.Book) error
    getByIDFn  func(ctx context.Context, id string) (model.Book, error)
    listFn     func(ctx context.Context, limit, offset int) ([]model.Book, error)
    updateFn   func(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error)
    deleteFn   func(ctx context.Context, id string) error
}

func (m *mockBookRepo) Create(ctx context.Context, b *model.Book) error {
    return m.createFn(ctx, b)
}

func (m *mockBookRepo) GetByID(ctx context.Context, id string) (model.Book, error) {
    return m.getByIDFn(ctx, id)
}

func (m *mockBookRepo) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
    return m.listFn(ctx, limit, offset)
}

func (m *mockBookRepo) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
    return m.updateFn(ctx, id, updates)
}

func (m *mockBookRepo) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

var _ repo.BookRepo = (*mockBookRepo)(nil)

func TestBookService_Create_Success(t *testing.T) {
    ctx := context.Background()

    mock := &mockBookRepo{
        createFn: func(_ context.Context, b *model.Book) error {
            b.ID = "book-1"
            b.Version = 1
            return nil
        },
    }

    svc := NewBookService(mock)
    book := &model.Book{Title: "Go Programming", Author: "Donovan"}
    err := svc.Create(ctx, book)

    require.NoError(t, err)
    require.NotEmpty(t, book.ID)
    require.Equal(t, 1, book.Version)
}

func TestBookService_GetByID_Success(t *testing.T) {
    ctx := context.Background()

    mock := &mockBookRepo{
        getByIDFn: func(_ context.Context, id string) (model.Book, error) {
            return model.Book{
                ID:            id,
                Title:         "Go Programming",
                Author:        "Donovan",
                Version:       1,
                PublishedYear: 2015,
            }, nil
        },
    }

    svc := NewBookService(mock)
    book, err := svc.GetByID(ctx, "book-1")

    require.NoError(t, err)
    require.Equal(t, "book-1", book.ID)
    require.Equal(t, "Go Programming", book.Title)
}

func TestBookService_GetByID_NotFound(t *testing.T) {
    ctx := context.Background()

    mock := &mockBookRepo{
        getByIDFn: func(_ context.Context, id string) (model.Book, error) {
            return model.Book{}, errors.New("not found")
        },
    }

    svc := NewBookService(mock)
    book, err := svc.GetByID(ctx, "nonexistent")

    require.Error(t, err)
    require.Equal(t, model.Book{}, book)
}

func TestBookService_Update_Success(t *testing.T) {
    ctx := context.Background()

    mock := &mockBookRepo{
        updateFn: func(_ context.Context, id string, updates map[string]interface{}) (*model.Book, error) {
            return &model.Book{
                ID:      id,
                Title:   "Go Programming - Updated",
                Author:  "Donovan",
                Version: 2,
            }, nil
        },
    }

    svc := NewBookService(mock)
    updates := map[string]interface{}{"title": "Go Programming - Updated"}
    book, err := svc.Update(ctx, "book-1", updates)

    require.NoError(t, err)
    require.Equal(t, "Go Programming - Updated", book.Title)
    require.Equal(t, 2, book.Version)
}

func TestBookService_List_Success(t *testing.T) {
    ctx := context.Background()

    mock := &mockBookRepo{
        listFn: func(_ context.Context, limit, offset int) ([]model.Book, error) {
            return []model.Book{
                {ID: "1", Title: "Book 1", Version: 1},
                {ID: "2", Title: "Book 2", Version: 1},
            }, nil
        },
    }

    svc := NewBookService(mock)
    books, err := svc.List(ctx, 10, 0)

    require.NoError(t, err)
    require.Len(t, books, 2)
}

func TestBookService_Delete_Success(t *testing.T) {
    ctx := context.Background()

    mock := &mockBookRepo{
        deleteFn: func(_ context.Context, id string) error {
            return nil
        },
    }

    svc := NewBookService(mock)
    err := svc.Delete(ctx, "book-1")

    require.NoError(t, err)
}