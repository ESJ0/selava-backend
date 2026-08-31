package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

// PedidoRepository es la interfaz que necesita el service para poder
// probarse con un fake, en vez de depender de *repository.PedidoRepository.
type PedidoRepository interface {
	Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error)
	UpdateEstado(ctx context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error)
}

type PedidoService struct {
	repo PedidoRepository
}

func NewPedidoService(repo PedidoRepository) *PedidoService {
	return &PedidoService{repo: repo}
}

func (s *PedidoService) CrearPedido(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error) {
	for i := range req.Prendas {
		if req.Prendas[i].Color != nil {
			value := strings.TrimSpace(*req.Prendas[i].Color)
			req.Prendas[i].Color = &value
		}
		if req.Prendas[i].Descripcion != nil {
			value := strings.TrimSpace(*req.Prendas[i].Descripcion)
			req.Prendas[i].Descripcion = &value
		}
	}
	if errs := validator.ValidatePedidoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req, usuarioID)
}

func (s *PedidoService) ActualizarEstado(ctx context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if usuarioID <= 0 {
		return nil, repository.ErrUsuarioNoEncontrado
	}
	if req.Observaciones != nil {
		value := strings.TrimSpace(*req.Observaciones)
		req.Observaciones = &value
	}
	if errs := validator.ValidatePedidoEstadoUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.UpdateEstado(ctx, pedidoID, req, usuarioID)
}
