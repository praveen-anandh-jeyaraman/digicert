package service

import (
    "context"
    "errors"
    "testing"

    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
    "github.com/stretchr/testify/require"
    "golang.org/x/crypto/bcrypt"
)

// Mock user repo
type mockUserRepo struct {
    createFn        func(ctx context.Context, u *model.User) error
    getByIDFn       func(ctx context.Context, id string) (*model.User, error)
    getByUsernameFn func(ctx context.Context, username string) (*model.User, error)
    getByEmailFn    func(ctx context.Context, email string) (*model.User, error)
    updateFn        func(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
    listFn          func(ctx context.Context, limit, offset int) ([]model.User, error)
    deleteFn        func(ctx context.Context, id string) error
}

func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error {
    return m.createFn(ctx, u)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
    return m.getByIDFn(ctx, id)
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return m.getByUsernameFn(ctx, username)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    return m.getByEmailFn(ctx, email)
}

func (m *mockUserRepo) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
    return m.updateFn(ctx, id, updates)
}

func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    return m.listFn(ctx, limit, offset)
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
    return m.deleteFn(ctx, id)
}

var _ repo.UserRepo = (*mockUserRepo)(nil)

func TestUserService_Register_Success(t *testing.T) {
    ctx := context.Background()
    mock := &mockUserRepo{
        getByEmailFn: func(_ context.Context, email string) (*model.User, error) {
            return nil, errors.New("not found")
        },
        createFn: func(_ context.Context, u *model.User) error {
            u.ID = "user-1"
            u.Role = "USER"
            return nil
        },
    }
    svc := NewUserService(mock)

    req := &model.RegisterRequest{
        Username: "john",
        Email:    "john@example.com",
        Password: "SecurePass123",
    }
    user, err := svc.Register(ctx, req)

    require.NoError(t, err)
    require.Equal(t, "john", user.Username)
    require.Equal(t, "USER", user.Role)
}

func TestUserService_ValidatePassword_Success(t *testing.T) {
    ctx := context.Background()
    // Create a valid bcrypt hash for "SecurePass123"
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SecurePass123"), bcrypt.DefaultCost)
    require.NoError(t, err)

    mock := &mockUserRepo{
        getByUsernameFn: func(_ context.Context, username string) (*model.User, error) {
            return &model.User{
                ID:       "user-1",
                Username: username,
                Password: string(hashedPassword),
                Role:     "USER",
            }, nil
        },
    }
    svc := NewUserService(mock)

    user, err := svc.ValidatePassword(ctx, "john", "SecurePass123")
    require.NoError(t, err)
    require.Equal(t, "john", user.Username)
}

func TestUserService_ValidatePassword_WrongPassword(t *testing.T) {
    ctx := context.Background()
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SecurePass123"), bcrypt.DefaultCost)
    require.NoError(t, err)

    mock := &mockUserRepo{
        getByUsernameFn: func(_ context.Context, username string) (*model.User, error) {
            return &model.User{
                ID:       "user-1",
                Username: username,
                Password: string(hashedPassword),
                Role:     "USER",
            }, nil
        },
    }
    svc := NewUserService(mock)

    user, err := svc.ValidatePassword(ctx, "john", "WrongPassword")
    require.Error(t, err)
    require.Nil(t, user)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
    ctx := context.Background()
    mock := &mockUserRepo{
        getByIDFn: func(_ context.Context, id string) (*model.User, error) {
            return nil, errors.New("not found")
        },
    }
    svc := NewUserService(mock)

    user, err := svc.GetByID(ctx, "nonexistent")
    require.Error(t, err)
    require.Nil(t, user)
}

func TestUserService_GetByID_Success(t *testing.T) {
    ctx := context.Background()
    mock := &mockUserRepo{
        getByIDFn: func(_ context.Context, id string) (*model.User, error) {
            return &model.User{
                ID:       id,
                Username: "john",
                Email:    "john@example.com",
                Role:     "USER",
            }, nil
        },
    }
    svc := NewUserService(mock)

    user, err := svc.GetByID(ctx, "user-1")
    require.NoError(t, err)
    require.Equal(t, "john", user.Username)
}

func TestUserService_List_Success(t *testing.T) {
    ctx := context.Background()
    mock := &mockUserRepo{
        listFn: func(_ context.Context, limit, offset int) ([]model.User, error) {
            return []model.User{
                {ID: "1", Username: "user1", Role: "USER"},
                {ID: "2", Username: "user2", Role: "ADMIN"},
            }, nil
        },
    }
    svc := NewUserService(mock)

    users, err := svc.List(ctx, 10, 0)
    require.NoError(t, err)
    require.Len(t, users, 2)
}