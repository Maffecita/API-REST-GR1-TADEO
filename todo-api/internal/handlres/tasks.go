package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todo-api/internal/middleware"
	"todo-api/internal/models"
	"todo-api/internal/repository"
)

// TaskHandler maneja todas las operaciones CRUD de tareas
type TaskHandler struct {
	repo *repository.SQLiteRepository
}

func NewTaskHandler(repo *repository.SQLiteRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// GetAll GET /api/tasks
func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	tasks, err := h.repo.GetAllTasks(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error al obtener tareas")
		return
	}

	// Evitar null en JSON cuando no hay tareas
	if tasks == nil {
		tasks = []models.Task{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// Create POST /api/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Prioridad por defecto
	if req.Priority == "" {
		req.Priority = models.PriorityMedium
	}

	task := &models.Task{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	if err := h.repo.CreateTask(task); err != nil {
		respondError(w, http.StatusInternalServerError, "error al crear la tarea")
		return
	}

	respondJSON(w, http.StatusCreated, task)
}

// GetByID GET /api/tasks/{id}
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	task, err := h.repo.GetTaskByID(id, userID)
	if err == models.ErrNotFound {
		respondError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error al obtener la tarea")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// Update PUT /api/tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// Verificar que la tarea existe y pertenece al usuario
	existing, err := h.repo.GetTaskByID(id, userID)
	if err == models.ErrNotFound {
		respondError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// Aplicar cambios (solo si se proporcionan)
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Priority != "" {
		existing.Priority = req.Priority
	}
	if req.DueDate != nil {
		existing.DueDate = req.DueDate
	}

	if err := h.repo.UpdateTask(existing); err != nil {
		respondError(w, http.StatusInternalServerError, "error al actualizar la tarea")
		return
	}

	respondJSON(w, http.StatusOK, existing)
}

// Delete DELETE /api/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.repo.DeleteTask(id, userID); err == models.ErrNotFound {
		respondError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "tarea eliminada correctamente"})
}

// MarkComplete PATCH /api/tasks/{id}/complete
func (h *TaskHandler) MarkComplete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	task, err := h.repo.MarkTaskComplete(id, userID)
	if err == models.ErrNotFound {
		respondError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
