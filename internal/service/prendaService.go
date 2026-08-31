package service

import (
	"context"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type PrendaRepository interface {
	CreateMany(ctx context.Context, pedidoID int, reqs []models.PrendaCreateRequest) ([]models.Prenda, error)
	AddServicio(ctx context.Context, prendaID, servicioID int) (*models.PrendaServicio, error)
	RemoveServicio(ctx context.Context, prendaID, servicioID int) error
}

type PrendaService struct {
	repo PrendaRepository
}

func NewPrendaService(repo PrendaRepository) *PrendaService {
	return &PrendaService{repo: repo}
}

func (s *PrendaService) RegistrarPrendas(ctx context.Context, pedidoID int, prendas []models.PrendaCreateRequest) ([]models.Prenda, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	validator.NormalizePrendasCreate(prendas)
	if errs := validator.ValidatePrendasCreate(prendas); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.CreateMany(ctx, pedidoID, prendas)
}

func (s *PrendaService) AsociarServicio(ctx context.Context, prendaID int, req *models.PrendaServicioCreateRequest) (*models.PrendaServicio, error) {
	if prendaID <= 0 {
		return nil, repository.ErrPrendaNoEncontrada
	}
	if req.ServicioID <= 0 {
		return nil, validator.ValidationErrors{{Field: "servicio_id", Message: "debe ser un entero positivo"}}
	}
	return s.repo.AddServicio(ctx, prendaID, req.ServicioID)
}

func (s *PrendaService) QuitarServicio(ctx context.Context, prendaID, servicioID int) error {
	if prendaID <= 0 {
		return repository.ErrPrendaNoEncontrada
	}
	if servicioID <= 0 {
		return repository.ErrPrendaServicioNoEncontrado
	}
	return s.repo.RemoveServicio(ctx, prendaID, servicioID)
}
