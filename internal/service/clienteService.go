package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

var ErrEmailEnUso = errors.New("el email ya está en uso por otro cliente")

type ClienteService struct {
	repo *repository.ClienteRepository
}

func NewClienteService(repo *repository.ClienteRepository) *ClienteService {
	return &ClienteService{repo: repo}
}

func (s *ClienteService) CrearCliente(ctx context.Context, req *models.ClienteCreateRequest) (*models.Cliente, error) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Apellido = strings.TrimSpace(req.Apellido)
	req.Email = strings.TrimSpace(req.Email)

	if errs := validator.ValidateClienteCreate(req); errs.HasErrors() {
		return nil, errs
	}

	if req.Email != "" {
		existe, err := s.repo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existe {
			return nil, ErrEmailEnUso
		}
	}

	cliente, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %w", err)
	}
	return cliente, nil
}

func (s *ClienteService) ObtenerCliente(ctx context.Context, id int) (*models.Cliente, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ClienteService) ListarClientes(ctx context.Context) ([]models.Cliente, error) {
	return s.repo.List(ctx)
}

func (s *ClienteService) ActualizarCliente(ctx context.Context, id int, req *models.ClienteUpdateRequest) (*models.Cliente, error) {

	if errs := validator.ValidateClienteUpdate(req); errs.HasErrors() {
		return nil, errs
	}

	if req.Email != nil && *req.Email != "" {
		existe, err := s.repo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if existe {
			actual, err := s.repo.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			if actual.Email != *req.Email {
				return nil, ErrEmailEnUso
			}
		}
	}
	return s.repo.Update(ctx, id, req)
}

func (s *ClienteService) EliminarCliente(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
