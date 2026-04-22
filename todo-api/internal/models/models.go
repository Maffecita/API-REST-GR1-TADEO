package models

import "time"

// ─── Prioridades válidas ──────────────────────────────────────────────────────
type Priority string

const (
	PriorityLow    Priority = "baja"
	PriorityMedium Priority = "media"
	PriorityHigh   Priority = "alta"
)

// ─── Task representa una tarea en el sistema ──────────────────────────────────
type Task struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	Completed   bool      `json:"completed"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTaskRequest es el cuerpo esperado al crear una tarea
type CreateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateTaskRequest es el cuerpo esperado al actualizar una tarea
type UpdateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// Validate valida los campos requeridos de una tarea
func (r *CreateTaskRequest) Validate() error {
	if r.Title == "" {
		return ErrEmptyTitle
	}
	if len(r.Title) > 150 {
		return ErrTitleTooLong
	}
	if r.Priority != "" && r.Priority != PriorityLow &&
		r.Priority != PriorityMedium && r.Priority != PriorityHigh {
		return ErrInvalidPriority
	}
	return nil
}

// ─── User representa un usuario del sistema ───────────────────────────────────
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Nunca se serializa en JSON
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest es el cuerpo esperado al registrar un usuario
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest es el cuerpo esperado al iniciar sesión
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse es la respuesta al autenticarse exitosamente
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Validate valida los campos del registro
func (r *RegisterRequest) Validate() error {
	if r.Username == "" || r.Email == "" || r.Password == "" {
		return ErrMissingFields
	}
	if len(r.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}
