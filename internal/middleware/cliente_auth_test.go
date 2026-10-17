package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
)

func TestAuthenticateClienteCargaIdentidad(t *testing.T) {
	const secret = "cliente-middleware-secret"
	token, err := auth.GenerateClienteToken(secret, 42)
	if err != nil {
		t.Fatal(err)
	}

	handler := NewAuthMiddleware(secret).AuthenticateCliente(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClienteClaimsFromContext(r.Context())
		if !ok || claims.ClienteID != 42 {
			t.Fatalf("identidad ausente o incorrecta: %#v", claims)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/cliente/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestAuthenticateClienteRechazaTokenDeUsuario(t *testing.T) {
	const secret = "cliente-middleware-secret"
	token, err := auth.GenerateToken(secret, 7, RolRecepcionista)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewAuthMiddleware(secret).AuthenticateCliente(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("el handler privado no debio ejecutarse")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/cliente/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestAuthenticateUsuarioRechazaTokenDeCliente(t *testing.T) {
	const secret = "cliente-middleware-secret"
	token, err := auth.GenerateClienteToken(secret, 42)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewAuthMiddleware(secret).Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("el handler administrativo no debio ejecutarse")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/clientes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
}
