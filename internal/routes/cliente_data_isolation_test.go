package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/middleware"
)

func TestIdentidadClienteNoPuedeSuplantarseConParametrosOCabeceras(t *testing.T) {
	router, _ := nuevoRouterClienteEnMemoria()
	clienteA := registrarClienteHTTP(t, router, "cliente-a@example.com")
	clienteB := registrarClienteHTTP(t, router, "cliente-b@example.com")
	tokenA := loginClienteHTTP(t, router, "cliente-a@example.com", "secreto-seguro")
	tokenB := loginClienteHTTP(t, router, "cliente-b@example.com", "secreto-seguro")

	casos := []struct {
		nombre   string
		token    string
		esperado int
		otro     int
	}{
		{"cliente A", tokenA, clienteA, clienteB},
		{"cliente B", tokenB, clienteB, clienteA},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/cliente/me?cliente_id=%d", caso.otro), nil)
			req.Header.Set("Authorization", "Bearer "+caso.token)
			req.Header.Set("X-Cliente-ID", fmt.Sprint(caso.otro))
			res := ejecutarRequest(router, req)
			if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), fmt.Sprintf(`"id":%d`, caso.esperado)) {
				t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
			}
			if strings.Contains(res.Body.String(), fmt.Sprintf(`"id":%d`, caso.otro)) {
				t.Fatalf("la respuesta uso una identidad proporcionada por el cliente: %s", res.Body.String())
			}
		})
	}
}

func TestTokenClienteNoAccedeADatosAdministrativos(t *testing.T) {
	router, _ := nuevoRouterClienteEnMemoria()
	registrarClienteHTTP(t, router, "cliente@example.com")
	token := loginClienteHTTP(t, router, "cliente@example.com", "secreto-seguro")

	for _, path := range []string{
		"/api/clientes/999",
		"/api/pedidos/999",
		"/api/pedidos/999/pagos",
		"/api/pedidos/999/saldo",
	} {
		t.Run(path, func(t *testing.T) {
			res := ejecutarJSON(router, http.MethodGet, path, "", token)
			if res.Code != http.StatusUnauthorized {
				t.Fatalf("ruta=%s status=%d body=%s", path, res.Code, res.Body.String())
			}
		})
	}
}

func TestTokenUsuarioInternoNoAccedeARutaPrivadaCliente(t *testing.T) {
	router, _ := nuevoRouterClienteEnMemoria()
	token, err := auth.GenerateToken(clienteSecuritySecret, 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatal(err)
	}
	res := ejecutarJSON(router, http.MethodGet, "/api/cliente/me", "", token)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func ejecutarRequest(router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}
