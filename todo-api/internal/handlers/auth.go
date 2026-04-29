package handlers

import (
	"encoding/json"
	"net/http"
	"todo-api/internal/auth"
	"todo-api/internal/models"
	"todo-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler maneja registro e inicio de sesión
type AuthHandler struct {
	repo       *repository.SQLiteRepository
	jwtService *auth.JWTService
}

func NewAuthHandler(repo *repository.SQLiteRepository, jwtSvc *auth.JWTService) *AuthHandler {
	return &AuthHandler{repo: repo, jwtService: jwtSvc}
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Hash de la contraseña
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error al procesar la contraseña")
		return
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
	}

	if err := h.repo.CreateUser(user, string(hash)); err != nil {
		if err == models.ErrDuplicate {
			respondError(w, http.StatusConflict, "el username o email ya están en uso")
			return
		}
		respondError(w, http.StatusInternalServerError, "error al crear usuario")
		return
	}

	token, _ := h.jwtService.GenerateToken(user.ID, user.Username)
	respondJSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: *user})
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	user, hash, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "credenciales incorrectas")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "credenciales incorrectas")
		return
	}

	token, _ := h.jwtService.GenerateToken(user.ID, user.Username)
	respondJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: *user})
}
