package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
)

// TipoPrendaRepository es la interfaz que necesita el service para poder
// probarse con un fake, en vez de depender de *repository.TipoPrendaRepository.
type TipoPrendaRepository interface {
	Create(ctx context.Context, req *models.TipoPrendaCreateRequest) (*models.TipoPrenda, error)
	GetByID(ctx context.Context, id int) (*models.TipoPrenda, error)
	List(ctx context.Context) ([]models.TipoPrenda, error)
	Update(ctx context.Context, id int, req *models.TipoPrendaUpdateRequest) (*models.TipoPrenda, error)
	Delete(ctx context.Context, id int) error
}

type TipoPrendaService struct {
	repo TipoPrendaRepository
}

func NewTipoPrendaService(repo TipoPrendaRepository) *TipoPrendaService {
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
