package service

import (
    "context"
    "errors"

    "golang.org/x/crypto/bcrypt"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
)

type UserService interface {
    RegisterAdmin(ctx context.Context, req *model.RegisterRequest) (*model.User, error) 
    Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
    GetByID(ctx context.Context, id string) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
    Delete(ctx context.Context, id string) error
    ValidatePassword(ctx context.Context, username, password string) (*model.User, error)
    List(ctx context.Context, limit, offset int) ([]model.User, error)
}

type userService struct {
    repo repo.UserRepo
}

func NewUserService(r repo.UserRepo) UserService {
    return &userService{repo: r}
}

func (s *userService) RegisterAdmin(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
    if req.Username == "" || req.Email == "" || req.Password == "" {
        return nil, errors.New("username, email, and password are required")
    }

    if len(req.Password) < 8 {
        return nil, errors.New("password must be at least 8 characters")
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

    u := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Role:     "admin",
    }

    if err := s.repo.Create(ctx, u); err != nil {
        return nil, err
    }

    u.Password = ""
    return u, nil
}
func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
    // Validate input
    if req.Username == "" || req.Email == "" || req.Password == "" {
        return nil, errors.New("username, email, and password are required")
    }

    if len(req.Password) < 8 {
        return nil, errors.New("password must be at least 8 characters")
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

    u := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Role:     "user",
    }

    if err := s.repo.Create(ctx, u); err != nil {
        return nil, err
    }

    u.Password = ""
    return u, nil
}

// GetByID retrieves a user by ID
func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
    return s.repo.GetByID(ctx, id)
}

// GetByUsername retrieves a user by username
func (s *userService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return s.repo.GetByUsername(ctx, username)
}

// GetByEmail retrieves a user by email
func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    return s.repo.GetByEmail(ctx, email)
}

// Update updates user information
func (s *userService) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
    delete(updates, "password_hash")
    delete(updates, "id")

    return s.repo.Update(ctx, id, updates)
}

func (s *userService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *userService) ValidatePassword(ctx context.Context, username, password string) (*model.User, error) {
    u, err := s.repo.GetByUsername(ctx, username)
    if err != nil {
        return nil, errors.New("invalid username or password")
    }

    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
        return nil, errors.New("invalid username or password")
    }

    u.Password = ""
    return u, nil
}

func (s *userService) List(ctx context.Context, limit, offset int) ([]model.User, error) {
    return s.repo.List(ctx, limit, offset)
}