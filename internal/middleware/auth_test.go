package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
)

func TestAuthenticateRejectsMissingToken(t *testing.T) {
	mw := NewAuthMiddleware("secret")
	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/clientes", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
	if res.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("expected WWW-Authenticate Bearer header, got %q", res.Header().Get("WWW-Authenticate"))
	}
}

func TestAuthenticateAddsClaimsToContext(t *testing.T) {
	const secret = "secret"
	token := mustToken(t, secret, 12, RolAdministrador)
	mw := NewAuthMiddleware(secret)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("expected claims in context")
		}
		if claims.UsuarioID != 12 || claims.RolID != RolAdministrador {
			t.Fatalf("unexpected claims: usuario_id=%d rol_id=%d", claims.UsuarioID, claims.RolID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/clientes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}

func TestRequireRolesRejectsDisallowedRole(t *testing.T) {
	const secret = "secret"
	token := mustToken(t, secret, 12, 99)
	mw := NewAuthMiddleware(secret)

	handler := mw.Authenticate(mw.RequireRoles(RolAdministrador)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/clientes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, res.Code)
	}
	if res.Header().Get("WWW-Authenticate") != "" {
		t.Fatalf("expected no WWW-Authenticate header for forbidden response, got %q", res.Header().Get("WWW-Authenticate"))
	}
}

func mustToken(t *testing.T, secret string, usuarioID, rolID int) string {
	t.Helper()
	token, err := auth.GenerateToken(secret, usuarioID, rolID)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	return token
}
