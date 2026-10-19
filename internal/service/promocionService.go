package service

import (
	"context"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
	"strings"
)

type PromocionRepository interface {
	List(context.Context) ([]models.Promocion, error)
	GetByID(context.Context, int) (*models.Promocion, error)
	Save(context.Context, int, *models.PromocionRequest) (*models.Promocion, error)
	SetActivo(context.Context, int, bool) (*models.Promocion, error)
	Cotizar(context.Context, *models.PedidoCreateRequest) (*models.PromocionCotizacion, error)
}
type PromocionService struct{ repo PromocionRepository }

func NewPromocionService(repo PromocionRepository) *PromocionService {
	return &PromocionService{repo: repo}
}
func (s *PromocionService) List(ctx context.Context) ([]models.Promocion, error) {
	return s.repo.List(ctx)
}
func (s *PromocionService) Get(ctx context.Context, id int) (*models.Promocion, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *PromocionService) Save(ctx context.Context, id int, req *models.PromocionRequest) (*models.Promocion, error) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Descripcion = strings.TrimSpace(req.Descripcion)
	if errs := validator.ValidatePromocion(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Save(ctx, id, req)
}
func (s *PromocionService) SetActivo(ctx context.Context, id int, activo bool) (*models.Promocion, error) {
	return s.repo.SetActivo(ctx, id, activo)
}
func (s *PromocionService) Cotizar(ctx context.Context, req *models.PedidoCreateRequest) (*models.PromocionCotizacion, error) {
	if errs := validator.ValidateCotizacion(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Cotizar(ctx, req)
}
