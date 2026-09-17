package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type MovimientoInventarioRepository interface {
	Create(ctx context.Context, req *models.MovimientoInventarioCreateRequest, usuarioID int) (*models.MovimientoInventario, error)
}

type MovimientoInventarioService struct {
	repo MovimientoInventarioRepository
}

func NewMovimientoInventarioService(repo MovimientoInventarioRepository) *MovimientoInventarioService {
	return &MovimientoInventarioService{repo: repo}
}

func (s *MovimientoInventarioService) RegistrarMovimiento(ctx context.Context, req *models.MovimientoInventarioCreateRequest, usuarioID int) (*models.MovimientoInventario, error) {
	if usuarioID <= 0 {
		return nil, repository.ErrUsuarioNoEncontrado
	}
	if req.Motivo != nil {
		motivo := strings.TrimSpace(*req.Motivo)
		if motivo == "" {
			req.Motivo = nil
		} else {
			req.Motivo = &motivo
		}
	}
	if errs := validator.ValidateMovimientoInventarioCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req, usuarioID)
}
