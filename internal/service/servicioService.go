package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type ServicioService struct {
	repo *repository.ServicioRepository
}

func NewServicioService(repo *repository.ServicioRepository) *ServicioService {
	return &ServicioService{repo: repo}
}

func (s *ServicioService) CrearServicio(ctx context.Context, req *models.ServicioCreateRequest) (*models.Servicio, error) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	if req.Descripcion != nil {
		descripcion := strings.TrimSpace(*req.Descripcion)
		req.Descripcion = &descripcion
	}
	if errs := validator.ValidateServicioCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req)
}

func (s *ServicioService) ObtenerServicio(ctx context.Context, id int) (*models.Servicio, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ServicioService) ListarServicios(ctx context.Context) ([]models.Servicio, error) {
	return s.repo.List(ctx)
}

func (s *ServicioService) ActualizarServicio(ctx context.Context, id int, req *models.ServicioUpdateRequest) (*models.Servicio, error) {
	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		req.Nombre = &nombre
	}
	if req.Descripcion != nil {
		descripcion := strings.TrimSpace(*req.Descripcion)
		req.Descripcion = &descripcion
	}
	if errs := validator.ValidateServicioUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Update(ctx, id, req)
}

func (s *ServicioService) EliminarServicio(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
