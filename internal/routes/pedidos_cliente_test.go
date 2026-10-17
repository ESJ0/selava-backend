package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

const pedidosClienteSecret = "pedidos-cliente-secret"

type authClienteRouteStub struct{}

func (authClienteRouteStub) Registrar(context.Context, *models.ClienteRegistroRequest) (*models.ClienteRegistrado, error) {
	return nil, nil
}
func (authClienteRouteStub) Login(context.Context, *models.ClienteLoginRequest) (string, error) {
	return "", nil
}
func (authClienteRouteStub) ObtenerPerfil(context.Context, int) (*models.ClienteResponse, error) {
	return &models.ClienteResponse{ID: 1, Activo: true}, nil
}

type pedidosClienteRouteRepo struct {
	service.PedidoRepository
	clienteID int
}

func (r *pedidosClienteRouteRepo) List(_ context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error) {
	r.clienteID = filter.ClienteID
	return &models.PedidoListResponse{Pedidos: []models.PedidoResumen{}, Pagina: filter.Pagina, Limite: filter.Limite}, nil
}

func (r *pedidosClienteRouteRepo) GetDetalleCliente(_ context.Context, pedidoID, clienteID int) (*models.PedidoDetalle, error) {
	r.clienteID = clienteID
	return &models.PedidoDetalle{Pedido: models.Pedido{ID: pedidoID, UsuarioID: 99}, Prendas: []models.Prenda{}, Pagos: []models.PagoDetalle{}}, nil
}

func (r *pedidosClienteRouteRepo) GetHistorialEstadosCliente(_ context.Context, pedidoID, clienteID int) ([]models.PedidoEstadoHistorialDetalle, error) {
	r.clienteID = clienteID
	return []models.PedidoEstadoHistorialDetalle{{
		PedidoEstadoHistorial: models.PedidoEstadoHistorial{ID: 1, PedidoID: pedidoID, UsuarioID: 99},
		Estado:                models.EstadoPedido{ID: 1, Nombre: "Recibido"},
		Usuario:               models.UsuarioResumen{ID: 99, Nombre: "Empleado"},
	}}, nil
}

func pedidosClienteRouter(repo *pedidosClienteRouteRepo) http.Handler {
	pedidoController := controller.NewPedidoController(service.NewPedidoService(repo))
	clienteAuthController := controller.NewClienteAuthController(authClienteRouteStub{})
	return NewRouterWithClienteAuth(nil, nil, nil, nil, nil, nil, pedidoController, nil, nil, nil,
		controller.NewAuthController(nil), clienteAuthController, pedidosClienteSecret, "http://localhost:5174")
}

func clienteRequest(t *testing.T, target string, clienteID int) *http.Request {
	t.Helper()
	token, err := auth.GenerateClienteToken(pedidosClienteSecret, clienteID)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestPedidosClienteUsanClienteIDDelToken(t *testing.T) {
	for _, target := range []string{
		"/api/cliente/pedidos",
		"/api/cliente/pedidos/7",
		"/api/cliente/pedidos/7/historial-estados",
	} {
		repo := &pedidosClienteRouteRepo{}
		res := httptest.NewRecorder()
		pedidosClienteRouter(repo).ServeHTTP(res, clienteRequest(t, target, 42))
		if res.Code != http.StatusOK || repo.clienteID != 42 {
			t.Fatalf("target=%s status=%d clienteID=%d body=%s", target, res.Code, repo.clienteID, res.Body.String())
		}
		if target != "/api/cliente/pedidos" && (containsJSONField(res.Body.String(), "usuario_id") || containsJSONField(res.Body.String(), "usuario")) {
			t.Fatalf("target=%s expuso datos internos: %s", target, res.Body.String())
		}
	}
}

func TestPedidosClienteRechazanTokenInternoYFaltaDeToken(t *testing.T) {
	router := pedidosClienteRouter(&pedidosClienteRouteRepo{})
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/api/cliente/pedidos", nil),
		func() *http.Request {
			token, err := auth.GenerateToken(pedidosClienteSecret, 7, middleware.RolAdministrador)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/cliente/pedidos", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			return req
		}(),
	} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
		}
	}
}

func containsJSONField(body, field string) bool {
	needle := `"` + field + `"`
	for i := 0; i+len(needle) <= len(body); i++ {
		if body[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
