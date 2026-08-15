package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type MetodoPagoService struct {
	repo *repository.MetodoPagoRepository
}

func NewMetodoPagoService(repo *repository.MetodoPagoRepository) *MetodoPagoService {
	return &MetodoPagoService{repo: repo}
}

func (s *MetodoPagoService) CrearMetodoPago(ctx context.Context, req *models.MetodoPagoCreateRequest) (*models.MetodoPago, error) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	if errs := validator.ValidateMetodoPagoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req)
}

func (s *MetodoPagoService) ObtenerMetodoPago(ctx context.Context, id int) (*models.MetodoPago, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MetodoPagoService) ListarMetodosPago(ctx context.Context) ([]models.MetodoPago, error) {
	return s.repo.List(ctx)
}

func (s *MetodoPagoService) ActualizarMetodoPago(ctx context.Context, id int, req *models.MetodoPagoUpdateRequest) (*models.MetodoPago, error) {
	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		req.Nombre = &nombre
	}
	if errs := validator.ValidateMetodoPagoUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Update(ctx, id, req)
}

func (s *MetodoPagoService) EliminarMetodoPago(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
