package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"todo-api/internal/auth"
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"
	"todo-api/internal/repository"
)

func main() {
	// ─── Configuración del puerto ────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ─── Inicializar repositorio (base de datos SQLite) ───────────────────
	repo, err := repository.NewSQLiteRepository("db/todo.db")
	if err != nil {
		log.Fatalf("Error al inicializar la base de datos: %v", err)
	}
	defer repo.Close()

	// ─── Inicializar servicios ────────────────────────────────────────────
	jwtService := auth.NewJWTService("mi-secreto-super-seguro-cambiar-en-produccion")

	// ─── Inicializar handlers ─────────────────────────────────────────────
	taskHandler := handlers.NewTaskHandler(repo)
	authHandler := handlers.NewAuthHandler(repo, jwtService)

	// ─── Router principal ─────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Rutas públicas (sin autenticación)
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/health", healthCheck)

	// Rutas protegidas (requieren JWT)
	protected := middleware.Chain(
		middleware.Logger,
		middleware.Auth(jwtService),
	)

	// Combined handler for /api/tasks (GET and POST)
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			protected(http.HandlerFunc(taskHandler.GetAll)).ServeHTTP(w, r)
		case http.MethodPost:
			protected(http.HandlerFunc(taskHandler.Create)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Combined handler for /api/tasks/{id} (GET, PUT, DELETE, PATCH)
	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path (e.g., /api/tasks/123 -> 123)
		id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		if id == "" || strings.Contains(id, "/") { // Invalid if empty or has extra slashes
			http.NotFound(w, r)
			return
		}

		// Dispatch based on method
		switch r.Method {
		case http.MethodGet:
			// Assuming GetByID expects ID; adjust if your handler uses context or params
			r.URL.Path = "/api/tasks/" + id // Restore path for handler if needed
			protected(http.HandlerFunc(taskHandler.GetByID)).ServeHTTP(w, r)
		case http.MethodPut:
			r.URL.Path = "/api/tasks/" + id
			protected(http.HandlerFunc(taskHandler.Update)).ServeHTTP(w, r)
		case http.MethodDelete:
			r.URL.Path = "/api/tasks/" + id
			protected(http.HandlerFunc(taskHandler.Delete)).ServeHTTP(w, r)
		case http.MethodPatch:
			r.URL.Path = "/api/tasks/" + id
			protected(http.HandlerFunc(taskHandler.MarkComplete)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ─── Iniciar servidor ─────────────────────────────────────────────────
	fmt.Printf("🚀 Servidor corriendo en http://localhost:%s\n", port)
	fmt.Println("📋 Endpoints disponibles:")
	fmt.Println("   POST   /api/auth/register")
	fmt.Println("   POST   /api/auth/login")
	fmt.Println("   GET    /api/tasks")
	fmt.Println("   POST   /api/tasks")
	fmt.Println("   GET    /api/tasks/{id}")
	fmt.Println("   PUT    /api/tasks/{id}")
	fmt.Println("   DELETE /api/tasks/{id}")
	fmt.Println("   PATCH  /api/tasks/{id}/complete")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","message":"API funcionando correctamente"}`)
}
