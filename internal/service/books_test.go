package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/praveen-anandh-jeyaraman/digicert/internal/model"
	"github.com/stretchr/testify/require"
)

// mockRepo is a tiny mock of repo.BookRepo used for unit tests.
type mockRepo struct {
	createFn func(ctx context.Context, b *model.Book) error
	getFn    func(ctx context.Context, id string) (model.Book, error)
	listFn   func(ctx context.Context, limit, offset int) ([]model.Book, error)
	updateFn func(ctx context.Context, b *model.Book) error
	deleteFn func(ctx context.Context, id string) error
}

func (m *mockRepo) Create(ctx context.Context, b *model.Book) error { return m.createFn(ctx, b) }
func (m *mockRepo) GetByID(ctx context.Context, id string) (model.Book, error) {
	return m.getFn(ctx, id)
}
func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
	return m.listFn(ctx, limit, offset)
}
func (m *mockRepo) Update(ctx context.Context, b *model.Book) error { return m.updateFn(ctx, b) }
func (m *mockRepo) Delete(ctx context.Context, id string) error     { return m.deleteFn(ctx, id) }

func TestBookService_Create_Success(t *testing.T) {
	ctx := context.Background()
	called := false
	mock := &mockRepo{
		createFn: func(_ context.Context, b *model.Book) error {
			called = true
			// simulate DB populating ID/version/timestamps
			b.ID = "fake-id-1"
			b.CreatedAt = time.Now().UTC()
			b.UpdatedAt = b.CreatedAt
			b.Version = 1
			return nil
		},
	}
	svc := NewBookService(mock)
	book := &model.Book{Title: "T", Author: "A"}
	err := svc.Create(ctx, book)
	require.NoError(t, err)
	require.True(t, called)
	require.NotEmpty(t, book.ID)
	require.Equal(t, 1, book.Version)
}

func TestBookService_Update_Conflict(t *testing.T) {
	ctx := context.Background()
	mock := &mockRepo{
		updateFn: func(_ context.Context, b *model.Book) error {
			return errors.New("conflict: stale version or not found")
		},
	}
	svc := NewBookService(mock)
	err := svc.Update(ctx, &model.Book{ID: "x", Version: 1})
	require.ErrorIs(t, err, ErrConflict)
}

func TestBookService_ListAndGet(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	mock := &mockRepo{
		listFn: func(_ context.Context, limit, offset int) ([]model.Book, error) {
			return []model.Book{
				{ID: "a", Title: "A", Author: "AA", CreatedAt: now, UpdatedAt: now, Version: 1},
			}, nil
		},
		getFn: func(_ context.Context, id string) (model.Book, error) {
			return model.Book{ID: id, Title: "A", Author: "AA", CreatedAt: now, UpdatedAt: now, Version: 1}, nil
		},
	}
	svc := NewBookService(mock)
	list, err := svc.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)

	got, err := svc.Get(ctx, "a")
	require.NoError(t, err)
	require.Equal(t, "a", got.ID)
}