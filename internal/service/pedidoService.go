package service

import (
	"context"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type PedidoService struct {
	repo *repository.PedidoRepository
}

func NewPedidoService(repo *repository.PedidoRepository) *PedidoService {
	return &PedidoService{repo: repo}
}

func (s *PedidoService) CrearPedido(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.Pedido, error) {
	if errs := validator.ValidatePedidoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req, usuarioID)
}
