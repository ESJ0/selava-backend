package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
)

type fakeEstadoPedidoRepository struct {
	estados []models.EstadoPedido
	err     error
}

func (r *fakeEstadoPedidoRepository) List(ctx context.Context) ([]models.EstadoPedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.estados, nil
}

func TestEstadoPedidoControllerListarReturnsEstados(t *testing.T) {
	repo := &fakeEstadoPedidoRepository{estados: []models.EstadoPedido{
		{ID: 1, Nombre: "Recibido", Orden: 1},
		{ID: 2, Nombre: "Rackeado", Orden: 2},
		{ID: 3, Nombre: "Entregado", Orden: 3},
		{ID: 4, Nombre: "Cancelado", Orden: 99},
	}}
	controller := NewEstadoPedidoController(servicelayer.NewEstadoPedidoService(repo))
	req := httptest.NewRequest(http.MethodGet, "/api/estados-pedido", nil)
	res := httptest.NewRecorder()

	controller.Listar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body []models.EstadoPedido
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 4 || body[1].Nombre != "Rackeado" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestEstadoPedidoControllerListarReturnsInternalErrorOnFailure(t *testing.T) {
	repo := &fakeEstadoPedidoRepository{err: errors.New("fallo de conexion")}
	controller := NewEstadoPedidoController(servicelayer.NewEstadoPedidoService(repo))
	req := httptest.NewRequest(http.MethodGet, "/api/estados-pedido", nil)
	res := httptest.NewRecorder()

	controller.Listar(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.Code)
	}
}
