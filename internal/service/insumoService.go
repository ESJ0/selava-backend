package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
)

// InsumoRepository es la interfaz que necesita el service para poder
// probarse con un fake, en vez de depender de *repository.InsumoRepository.
type InsumoRepository interface {
	Create(ctx context.Context, req *models.InsumoCreateRequest) (*models.Insumo, error)
	GetByID(ctx context.Context, id int) (*models.Insumo, error)
	List(ctx context.Context) ([]models.Insumo, error)
	ListLowStock(ctx context.Context) ([]models.Insumo, error)
	Update(ctx context.Context, id int, req *models.InsumoUpdateRequest) (*models.Insumo, error)
	Delete(ctx context.Context, id int) error
}

type InsumoService struct {
	repo InsumoRepository
}

func NewInsumoService(repo InsumoRepository) *InsumoService {
	return &InsumoService{repo: repo}
}

func (s *InsumoService) CrearInsumo(ctx context.Context, req *models.InsumoCreateRequest) (*models.Insumo, error) {
	normalizarInsumoCreate(req)
	if errs := validator.ValidateInsumoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req)
}

func (s *InsumoService) ObtenerInsumo(ctx context.Context, id int) (*models.Insumo, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *InsumoService) ListarInsumos(ctx context.Context) ([]models.Insumo, error) {
	return s.repo.List(ctx)
}

func (s *InsumoService) ListarInsumosBajoStock(ctx context.Context) ([]models.Insumo, error) {
	return s.repo.ListLowStock(ctx)
}

func (s *InsumoService) ActualizarInsumo(ctx context.Context, id int, req *models.InsumoUpdateRequest) (*models.Insumo, error) {
	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		req.Nombre = &nombre
	}
	if req.Descripcion != nil {
		descripcion := strings.TrimSpace(*req.Descripcion)
		req.Descripcion = &descripcion
	}
	if req.UnidadMedida != nil {
		unidad := strings.TrimSpace(*req.UnidadMedida)
		req.UnidadMedida = &unidad
	}
	if errs := validator.ValidateInsumoUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Update(ctx, id, req)
}

func (s *InsumoService) EliminarInsumo(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func normalizarInsumoCreate(req *models.InsumoCreateRequest) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.UnidadMedida = strings.TrimSpace(req.UnidadMedida)
	if req.Descripcion != nil {
		descripcion := strings.TrimSpace(*req.Descripcion)
		req.Descripcion = &descripcion
	}
}
