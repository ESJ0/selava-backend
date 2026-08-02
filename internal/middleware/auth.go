package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ESJ0/selava-backend/internal/auth"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

const (
	RolAdministrador = 1
	RolEmpleado      = 2
)

type AuthMiddleware struct {
	jwtSecret string
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
			return
		}

		claims, err := auth.ValidateToken(m.jwtSecret, tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "token invalido o expirado")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireRoles(allowedRoles ...int) func(http.Handler) http.Handler {
	allowed := make(map[int]struct{}, len(allowedRoles))
	for _, rolID := range allowedRoles {
		allowed[rolID] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
				return
			}

			if len(allowed) > 0 {
				if _, ok := allowed[claims.RolID]; !ok {
					respondError(w, http.StatusForbidden, "no tienes permisos para acceder a este recurso")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*auth.Claims)
	return claims, ok
}

func bearerToken(authHeader string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(authHeader))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", false
	}
	return fields[1], true
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}
