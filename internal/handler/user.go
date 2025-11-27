package handler

import (
    "encoding/json"
    "log"
    "net/http"    
    "strconv"
    "strings"
    "context"

    "github.com/go-chi/chi/v5"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/logger"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/model"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
)

type UserHandler struct {
    userSvc service.UserService
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
    return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) RegisterAdmin(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    var req model.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    user, err := h.userSvc.RegisterAdmin(r.Context(), &req)
    if err != nil {
        log.Printf("[%s] Admin registration failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to register admin")
        return
    }

    // Track metric
    cwLogger := logger.GetLogger()
    if cwLogger != nil {
        cwLogger.PutMetric(r.Context(), "AdminRegistered", 1, "Count")
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(user)
    log.Printf("[%s] Admin registered: %s", requestID, user.Username)
}
// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         Auth
// @Accept       json
// @Param        request  body      model.RegisterRequest  true  "Registration data"
// @Produce      json
// @Success      201  {object}  model.RegisterResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /auth/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    var req model.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate input
    errs := ValidationErrors{}
    req.Username = strings.TrimSpace(req.Username)
    req.Email = strings.TrimSpace(req.Email)
    req.Password = strings.TrimSpace(req.Password)

    if req.Username == "" {
        errs["username"] = "username is required"
    } else if len(req.Username) < 3 {
        errs["username"] = "username must be at least 3 characters"
    } else if len(req.Username) > 50 {
        errs["username"] = "username must be at most 50 characters"
    }

    if req.Email == "" {
        errs["email"] = "email is required"
    } else if !isValidEmail(req.Email) {
        errs["email"] = "invalid email format"
    }

    if req.Password == "" {
        errs["password"] = "password is required"
    } else if len(req.Password) < 8 {
        errs["password"] = "password must be at least 8 characters"
    }

    if len(errs) > 0 {
        log.Printf("[%s] Validation failed: %v", requestID, errs)
        WriteValidationErrors(r.Context(), w, errs)
        return
    }

    user, err := h.userSvc.Register(r.Context(), &req)
    if err != nil {
        if strings.Contains(err.Error(), "already exists") {
            log.Printf("[%s] Registration failed: %v", requestID, err)
            WriteError(r.Context(), w, http.StatusConflict, err.Error())
            return
        }
        log.Printf("[%s] Registration failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to register user")
        return
    }

    resp := model.RegisterResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(resp)
    log.Printf("[%s] User registered successfully: %s", requestID, user.ID)
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get current user profile
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  model.User
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    userID := GetUserID(r.Context())

    if userID == "" {
        log.Printf("[%s] Unauthorized", requestID)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    user, err := h.userSvc.GetByID(r.Context(), userID)
    if err != nil {
        log.Printf("[%s] User not found: %s", requestID, userID)
        WriteError(r.Context(), w, http.StatusNotFound, "User not found")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(user)
    log.Printf("[%s] User profile retrieved: %s", requestID, userID)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update current user profile
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Param        request  body      model.UpdateUserRequest  true  "Update data"
// @Produce      json
// @Success      200  {object}  model.User
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /users/me [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    userID := GetUserID(r.Context())

    if userID == "" {
        log.Printf("[%s] Unauthorized", requestID)
        WriteError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req model.UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[%s] Invalid request: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusBadRequest, "Invalid request body")
        return
    }

    errs := ValidationErrors{}
    if req.Email != "" && !isValidEmail(req.Email) {
        errs["email"] = "invalid email format"
    }

    if len(errs) > 0 {
        WriteValidationErrors(r.Context(), w, errs)
        return
    }

    updates := map[string]interface{}{}
    if req.Email != "" {
        updates["email"] = req.Email
    }

    if len(updates) == 0 {
        WriteError(r.Context(), w, http.StatusBadRequest, "No fields to update")
        return
    }

    user, err := h.userSvc.Update(r.Context(), userID, updates)
    if err != nil {
        if strings.Contains(err.Error(), "already exists") {
            log.Printf("[%s] Update failed: %v", requestID, err)
            WriteError(r.Context(), w, http.StatusConflict, "Email already in use")
            return
        }
        log.Printf("[%s] Update failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to update profile")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(user)
    log.Printf("[%s] User profile updated: %s", requestID, userID)
}
// ListUsers godoc
// @Summary      List all users (admin)
// @Description  Get all users in the system
// @Tags         Admin
// @Security     BearerAuth
// @Param        limit   query     int     false  "Items per page"  default(20)
// @Param        offset  query     int     false  "Pagination offset"  default(0)
// @Produce      json
// @Success      200  {array}   model.User
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /admin/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())

    limit := 20
    offset := 0

    if l := r.URL.Query().Get("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
            limit = parsed
        }
    }

    if o := r.URL.Query().Get("offset"); o != "" {
        if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
            offset = parsed
        }
    }

    users, err := h.userSvc.List(r.Context(), limit, offset)
    if err != nil {
        log.Printf("[%s] List users failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to list users")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(users)
    log.Printf("[%s] Listed %d users", requestID, len(users))
}

// GetUser godoc
// @Summary      Get user details (admin)
// @Description  Get a specific user by ID
// @Tags         Admin
// @Security     BearerAuth
// @Param        id   path  string  true  "User ID"
// @Produce      json
// @Success      200  {object}  model.User
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /admin/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    user, err := h.userSvc.GetByID(r.Context(), id)
    if err != nil {
        log.Printf("[%s] User not found: %s", requestID, id)
        WriteError(r.Context(), w, http.StatusNotFound, "User not found")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(user)
}

// DeleteUser godoc
// @Summary      Delete user (admin)
// @Description  Delete a user by ID
// @Tags         Admin
// @Security     BearerAuth
// @Param        id   path  string  true  "User ID"
// @Success      204
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    if err := h.userSvc.Delete(r.Context(), id); err != nil {
        log.Printf("[%s] Delete failed: %v", requestID, err)
        WriteError(r.Context(), w, http.StatusInternalServerError, "Failed to delete user")
        return
    }

    w.WriteHeader(http.StatusNoContent)
    log.Printf("[%s] User deleted: %s", requestID, id)
}

func isValidEmail(email string) bool {
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
func GetUserID(ctx context.Context) string {
    userID, ok := ctx.Value("user_id").(string)
    if !ok {
        return ""
    }
    return userID
}