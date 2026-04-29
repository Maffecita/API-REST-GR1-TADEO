package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
	"todo-api/internal/auth"
)

// contextKey es el tipo para las claves del contexto (evita colisiones)
type contextKey string

const UserIDKey contextKey = "userID"
const UsernameKey contextKey = "username"

// Middleware es el tipo de una función middleware
type Middleware func(http.Handler) http.Handler

// Chain encadena múltiples middlewares de afuera hacia adentro
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Logger registra cada petición con método, ruta y duración
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("→ %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("← %s %s [%v]", r.Method, r.URL.Path, time.Since(start))
	})
}

// Auth valida el JWT en el header Authorization
func Auth(jwtSvc *auth.JWTService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"token requerido"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := jwtSvc.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"token inválido o expirado"}`, http.StatusUnauthorized)
				return
			}

			// Inyectar datos del usuario en el contexto
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extrae el userID del contexto de la petición
func GetUserID(r *http.Request) int {
	id, _ := r.Context().Value(UserIDKey).(int)
	return id
}
