package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type PagoRepository interface {
	Create(ctx context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error)
}

type PagoService struct{ repo PagoRepository }

func NewPagoService(repo PagoRepository) *PagoService { return &PagoService{repo: repo} }

func (s *PagoService) RegistrarPago(ctx context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if usuarioID <= 0 {
		return nil, repository.ErrUsuarioNoEncontrado
	}
	if req.Referencia != nil {
		referencia := strings.TrimSpace(*req.Referencia)
		if referencia == "" {
			req.Referencia = nil
		} else {
			req.Referencia = &referencia
		}
	}
	if errs := validator.ValidatePagoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, pedidoID, req, usuarioID)
}
