package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/models"
)

type fakeClienteAuthRoutesService struct{}

func (fakeClienteAuthRoutesService) Registrar(context.Context, *models.ClienteRegistroRequest) (*models.ClienteRegistrado, error) {
	return &models.ClienteRegistrado{
		Cliente: models.Cliente{ID: 42, Nombre: "Ana", Activo: true},
		Cuenta:  models.CuentaCliente{ID: 8, ClienteID: 42, Email: "ana@example.com", Activo: true},
	}, nil
}

func (fakeClienteAuthRoutesService) Login(context.Context, *models.ClienteLoginRequest) (string, error) {
	return "jwt-cliente", nil
}

func clienteAuthRouter() http.Handler {
	clienteAuthController := controller.NewClienteAuthController(fakeClienteAuthRoutesService{})
	return NewRouterWithClienteAuth(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, clienteAuthController, "route-client-secret", "")
}

func TestClienteAuthRoutesExponeRegistroYLoginPublicos(t *testing.T) {
	casos := []struct {
		ruta string
		body string
	}{
		{"/api/auth/clientes/registro", `{"nombre":"Ana"}`},
		{"/api/auth/clientes/login", `{"email":"ana@example.com","password":"secreto"}`},
	}
	for _, caso := range casos {
		t.Run(caso.ruta, func(t *testing.T) {
			res := httptest.NewRecorder()
			clienteAuthRouter().ServeHTTP(res, httptest.NewRequest(http.MethodPost, caso.ruta, strings.NewReader(caso.body)))
			if res.Code == http.StatusNotFound || res.Code == http.StatusUnauthorized {
				t.Fatalf("ruta publica no disponible: status=%d body=%s", res.Code, res.Body.String())
			}
		})
	}
}

func TestClienteAuthRoutesProtegeMeYCargaCliente(t *testing.T) {
	router := clienteAuthRouter()
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/cliente/me", nil))
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("sin token status = %d", res.Code)
	}

	token, err := auth.GenerateClienteToken("route-client-secret", 42)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/cliente/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"cliente_id":42`) {
		t.Fatalf("con token status=%d body=%s", res.Code, res.Body.String())
	}
}
