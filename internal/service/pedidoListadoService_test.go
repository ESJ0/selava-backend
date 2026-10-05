package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

type listadoServiceRepo struct {
	PedidoRepository
	calls  int
	filter models.PedidoListFilter
	err    error
}

func (r *listadoServiceRepo) List(_ context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error) {
	r.calls++
	r.filter = filter
	return &models.PedidoListResponse{Pedidos: []models.PedidoResumen{}, Pagina: filter.Pagina, Limite: filter.Limite}, r.err
}

func TestListarPedidosValidatesBeforeRepository(t *testing.T) {
	repo := &listadoServiceRepo{}
	service := NewPedidoService(repo)
	if _, err := service.ListarPedidos(context.Background(), models.PedidoListRequest{Limite: "101"}); err == nil || repo.calls != 0 {
		t.Fatal("la consulta inválida alcanzó el repositorio")
	}
	result, err := service.ListarPedidos(context.Background(), models.PedidoListRequest{Buscar: "  Ana  "})
	if err != nil || repo.calls != 1 || repo.filter.Buscar != "Ana" || result.Pagina != 1 || result.Limite != 20 {
		t.Fatalf("result=%+v filter=%+v err=%v", result, repo.filter, err)
	}
}

func TestListarPedidosPropagatesRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	_, err := NewPedidoService(&listadoServiceRepo{err: want}).ListarPedidos(context.Background(), models.PedidoListRequest{})
	if !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}
