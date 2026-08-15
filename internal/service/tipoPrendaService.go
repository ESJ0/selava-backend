package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type TipoPrendaService struct {
	repo *repository.TipoPrendaRepository
}

func NewTipoPrendaService(repo *repository.TipoPrendaRepository) *TipoPrendaService {
	return &TipoPrendaService{repo: repo}
}

func (s *TipoPrendaService) CrearTipoPrenda(ctx context.Context, req *models.TipoPrendaCreateRequest) (*models.TipoPrenda, error) {
	normalizarTipoPrendaCreate(req)
	if errs := validator.ValidateTipoPrendaCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req)
}

func (s *TipoPrendaService) ObtenerTipoPrenda(ctx context.Context, id int) (*models.TipoPrenda, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TipoPrendaService) ListarTiposPrenda(ctx context.Context) ([]models.TipoPrenda, error) {
	return s.repo.List(ctx)
}

func (s *TipoPrendaService) ActualizarTipoPrenda(ctx context.Context, id int, req *models.TipoPrendaUpdateRequest) (*models.TipoPrenda, error) {
	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		req.Nombre = &nombre
	}
	if req.Descripcion != nil {
		descripcion := strings.TrimSpace(*req.Descripcion)
		req.Descripcion = &descripcion
	}
	if errs := validator.ValidateTipoPrendaUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Update(ctx, id, req)
}

func (s *TipoPrendaService) EliminarTipoPrenda(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func normalizarTipoPrendaCreate(req *models.TipoPrendaCreateRequest) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	if req.Descripcion != nil {
		descripcion := strings.TrimSpace(*req.Descripcion)
		req.Descripcion = &descripcion
	}
}
