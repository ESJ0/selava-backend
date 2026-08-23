package service

import (
	"context"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
)

// PedidoRepository es la interfaz que necesita el service para poder
// probarse con un fake, en vez de depender de *repository.PedidoRepository.
type PedidoRepository interface {
	Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error)
}

type PedidoService struct {
	repo PedidoRepository
}

func NewPedidoService(repo PedidoRepository) *PedidoService {
	return &PedidoService{repo: repo}
}

func (s *PedidoService) CrearPedido(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error) {
	if errs := validator.ValidatePedidoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req, usuarioID)
}
