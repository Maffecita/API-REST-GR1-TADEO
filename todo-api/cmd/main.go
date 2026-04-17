package main


import (
  "fmt"
  "log"
  "net/http"
  "os"
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
  mux.HandleFunc("POST /api/auth/register", authHandler.Register)
  mux.HandleFunc("POST /api/auth/login", authHandler.Login)
  mux.HandleFunc("GET /api/health", healthCheck)

 

  // Rutas protegidas (requieren JWT)
  protected := middleware.Chain(
    middleware.Logger,
    middleware.Auth(jwtService),
  )

 

  mux.Handle("GET /api/tasks", protected(http.HandlerFunc(taskHandler.GetAll)))
  mux.Handle("POST /api/tasks", protected(http.HandlerFunc(taskHandler.Create)))
  mux.Handle("GET /api/tasks/{id}", protected(http.HandlerFunc(taskHandler.GetByID)))
  mux.Handle("PUT /api/tasks/{id}", protected(http.HandlerFunc(taskHandler.Update)))
  mux.Handle("DELETE /api/tasks/{id}", protected(http.HandlerFunc(taskHandler.Delete)))
  mux.Handle("PATCH /api/tasks/{id}/complete", protected(http.HandlerFunc(taskHandler.MarkComplete)))

 

  // ─── Iniciar servidor ─────────────────────────────────────────────────
  fmt.Printf("🚀 Servidor corriendo en http://localhost:%s\n", port)
  fmt.Println("📋 Endpoints disponibles:")
  fmt.Println("   POST   /api/auth/register")
  fmt.Println("   POST   /api/auth/login")
  fmt.Println("   GET    /api/tasks")
  fmt.Println("   POST   /api/tasks")
  fmt.Println("   GET    /api/tasks/{id}")
  fmt.Println("   PUT    /api/tasks/{id}")
  fmt.Println("   DELETE /api/tasks/{id}")
  fmt.Println("   PATCH  /api/tasks/{id}/complete")

 

  if err := http.ListenAndServe(":"+port, mux); err != nil {
    log.Fatalf("Error al iniciar el servidor: %v", err)
  }
}

 

func healthCheck(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  fmt.Fprint(w, `{"status":"ok","message":"API funcionando correctamente"}`)
}