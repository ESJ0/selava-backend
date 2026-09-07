package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
)

func TestCancelarPedidoRequiereAutenticacionYRolAutorizado(t *testing.T) {
	const secret = "secret-de-prueba-sprint-3"
	middleware := NewAuthMiddleware(secret)
	handler := middleware.Authenticate(middleware.RequireRoles(RolAdministrador, RolRecepcionista)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	for _, tc := range []struct {
		name      string
		role      int
		withToken bool
		want      int
	}{
		{"sin token", 0, false, http.StatusUnauthorized},
		{"operario", RolOperario, true, http.StatusForbidden},
		{"administrador", RolAdministrador, true, http.StatusNoContent},
		{"recepcionista", RolRecepcionista, true, http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/pedidos/1/cancelar", nil)
			if tc.withToken {
				token, err := auth.GenerateToken(secret, 9, tc.role)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("status=%d, esperado=%d: %s", res.Code, tc.want, res.Body.String())
			}
		})
	}
}
