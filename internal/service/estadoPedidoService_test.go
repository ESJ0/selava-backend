package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

type fakeEstadoPedidoRepo struct {
	estados []models.EstadoPedido
	err     error
}

func (r *fakeEstadoPedidoRepo) List(ctx context.Context) ([]models.EstadoPedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.estados, nil
}

func TestEstadoPedidoServiceListarEstados(t *testing.T) {
	repo := &fakeEstadoPedidoRepo{estados: []models.EstadoPedido{
		{ID: 1, Nombre: "Recibido", Orden: 1},
		{ID: 2, Nombre: "Rackeado", Orden: 2},
		{ID: 3, Nombre: "Entregado", Orden: 3},
		{ID: 4, Nombre: "Cancelado", Orden: 99},
	}}
	service := NewEstadoPedidoService(repo)

	estados, err := service.ListarEstados(context.Background())
	if err != nil {
		t.Fatalf("ListarEstados returned error: %v", err)
	}
	if len(estados) != 4 {
		t.Fatalf("expected 4 estados, got %d", len(estados))
	}
	if estados[0].Nombre != "Recibido" || estados[3].Nombre != "Cancelado" {
		t.Fatalf("unexpected order: %+v", estados)
	}
}

func TestEstadoPedidoServiceListarEstadosPropagatesRepositoryError(t *testing.T) {
	repo := &fakeEstadoPedidoRepo{err: errors.New("fallo de conexion")}
	service := NewEstadoPedidoService(repo)

	_, err := service.ListarEstados(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
