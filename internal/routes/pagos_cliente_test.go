package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

type pagosClienteRouteRepo struct {
	service.PagoRepository
	clienteID int
}

func (r *pagosClienteRouteRepo) ListByPedidoCliente(_ context.Context, pedidoID, clienteID int) ([]models.PagoDetalle, error) {
	r.clienteID = clienteID
	return []models.PagoDetalle{{
		Pago:       models.Pago{ID: 1, PedidoID: pedidoID, UsuarioID: 99, Monto: 25},
		MetodoPago: models.MetodoPago{ID: 1, Nombre: "Efectivo"},
		Usuario:    &models.PagoUsuario{ID: 99, Nombre: "Empleado"},
	}}, nil
}

func (r *pagosClienteRouteRepo) GetSaldoCliente(_ context.Context, pedidoID, clienteID int) (*models.SaldoPedido, error) {
	r.clienteID = clienteID
	return &models.SaldoPedido{PedidoID: pedidoID, Total: 75, TotalPagado: 25, SaldoPendiente: 50}, nil
}

func pagosClienteRouter(repo *pagosClienteRouteRepo) http.Handler {
	pagoController := controller.NewPagoController(service.NewPagoService(repo))
	clienteAuthController := controller.NewClienteAuthController(authClienteRouteStub{})
	return NewRouterWithClienteAuth(nil, nil, nil, nil, nil, nil, nil, pagoController, nil, nil,
		controller.NewAuthController(nil), clienteAuthController, pedidosClienteSecret, "http://localhost:5174")
}

func TestPagosClienteUsanClienteIDDelTokenYOcultanEmpleado(t *testing.T) {
	for _, target := range []string{
		"/api/cliente/pedidos/7/pagos",
		"/api/cliente/pedidos/7/saldo",
	} {
		repo := &pagosClienteRouteRepo{}
		res := httptest.NewRecorder()
		pagosClienteRouter(repo).ServeHTTP(res, clienteRequest(t, target, 42))
		if res.Code != http.StatusOK || repo.clienteID != 42 {
			t.Fatalf("target=%s status=%d clienteID=%d body=%s", target, res.Code, repo.clienteID, res.Body.String())
		}
		if strings.Contains(res.Body.String(), "usuario_id") || strings.Contains(res.Body.String(), `"usuario"`) || strings.Contains(res.Body.String(), "Empleado") {
			t.Fatalf("target=%s expuso datos internos: %s", target, res.Body.String())
		}
	}
}

func TestPagosClienteRechazanTokenInternoYFaltaDeToken(t *testing.T) {
	router := pagosClienteRouter(&pagosClienteRouteRepo{})
	for _, target := range []string{"/api/cliente/pedidos/7/pagos", "/api/cliente/pedidos/7/saldo"} {
		requests := []*http.Request{httptest.NewRequest(http.MethodGet, target, nil)}
		token, err := auth.GenerateToken(pedidosClienteSecret, 7, middleware.RolAdministrador)
		if err != nil {
			t.Fatal(err)
		}
		internalRequest := httptest.NewRequest(http.MethodGet, target, nil)
		internalRequest.Header.Set("Authorization", "Bearer "+token)
		requests = append(requests, internalRequest)
		for _, request := range requests {
			res := httptest.NewRecorder()
			router.ServeHTTP(res, request)
			if res.Code != http.StatusUnauthorized {
				t.Fatalf("target=%s status=%d body=%s", target, res.Code, res.Body.String())
			}
		}
	}
}
