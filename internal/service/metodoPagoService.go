package service

import (
	"context"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
	"strings"
)

type MetodoPagoRepository interface {
	Create(context.Context, *models.MetodoPagoCreateRequest) (*models.MetodoPago, error)
	GetByID(context.Context, int) (*models.MetodoPago, error)
	List(context.Context) ([]models.MetodoPago, error)
	Update(context.Context, int, *models.MetodoPagoUpdateRequest) (*models.MetodoPago, error)
	Delete(context.Context, int) error
}
type MetodoPagoService struct{ repo MetodoPagoRepository }

func NewMetodoPagoService(repo MetodoPagoRepository) *MetodoPagoService {
	return &MetodoPagoService{repo: repo}
}
func (s *MetodoPagoService) CrearMetodoPago(ctx context.Context, req *models.MetodoPagoCreateRequest) (*models.MetodoPago, error) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	if e := validator.ValidateMetodoPagoCreate(req); e.HasErrors() {
		return nil, e
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
		v := strings.TrimSpace(*req.Nombre)
		req.Nombre = &v
	}
	if e := validator.ValidateMetodoPagoUpdate(req); e.HasErrors() {
		return nil, e
	}
	return s.repo.Update(ctx, id, req)
}
func (s *MetodoPagoService) EliminarMetodoPago(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
