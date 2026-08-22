package service

import (
	"context"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type PrendaService struct {
	repo *repository.PrendaRepository
}

func NewPrendaService(repo *repository.PrendaRepository) *PrendaService {
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
