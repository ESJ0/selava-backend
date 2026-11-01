package service

import (
	"context"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
	"sort"
	"strconv"
	"strings"
	"time"
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

// ListPublicas devuelve solo las promociones activas y vigentes en "now", de la más nueva a la más antigua,
// sin repetir registros idénticos (mismo nombre, prenda, medida y precio).
func (s *PromocionService) ListPublicas(ctx context.Context, now time.Time) ([]models.PromocionPublica, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]models.PromocionPublica, 0, len(items))
	seen := map[string]bool{}
	for _, p := range items {
		if !p.Activo || now.Before(p.FechaInicio) || now.After(p.FechaFin) {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(p.Nombre)) + "|" + strconv.Itoa(p.TipoPrendaID) + "|" + strconv.FormatFloat(p.MedidaRequerida, 'f', -1, 64) + "|" + strconv.FormatFloat(p.PrecioPromocional, 'f', -1, 64)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, models.PromocionPublica{ID: p.ID, Nombre: p.Nombre, Descripcion: p.Descripcion, TipoPrendaID: p.TipoPrendaID, TipoPrendaNombre: p.TipoPrendaNombre, ModalidadCobro: p.ModalidadCobro, MedidaRequerida: p.MedidaRequerida, PrecioPromocional: p.PrecioPromocional, FechaInicio: p.FechaInicio, FechaFin: p.FechaFin})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].ID > result[j].ID })
	return result, nil
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
