package service

import (
	"context"

	"github.com/ESJ0/selava-backend/internal/models"
)

// EstadoPedidoRepository es la interfaz que necesita el service para poder
// probarse con un fake, en vez de depender de *repository.EstadoPedidoRepository.
type EstadoPedidoRepository interface {
	List(ctx context.Context) ([]models.EstadoPedido, error)
}

type EstadoPedidoService struct {
	repo EstadoPedidoRepository
}

func NewEstadoPedidoService(repo EstadoPedidoRepository) *EstadoPedidoService {
	return &EstadoPedidoService{repo: repo}
}

func (s *EstadoPedidoService) ListarEstados(ctx context.Context) ([]models.EstadoPedido, error) {
	return s.repo.List(ctx)
}
