package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type UsuarioService struct{ repo *repository.UsuarioRepository }

func NewUsuarioService(repo *repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: repo}
}

func (s *UsuarioService) Listar(ctx context.Context) ([]models.UsuarioResponse, error) {
	return s.repo.List(ctx)
}

func (s *UsuarioService) Crear(ctx context.Context, req *models.UsuarioCreateRequest) (*models.UsuarioResponse, error) {
	normalizarUsuarioCreate(req)
	if req.RolID == 1 {
		if req.Nombre == "" {
			req.Nombre = "Administrador"
		}
		if req.Apellido == "" {
			req.Apellido = "SeLava"
		}
	}
	if errs := validator.ValidateUsuarioCreate(req); errs.HasErrors() {
		return nil, errs
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("protegiendo password de usuario: %w", err)
	}
	usuario, err := s.repo.Create(ctx, req, hash)
	if err != nil {
		return nil, err
	}
	response := usuario.ToResponse()
	return &response, nil
}

func (s *UsuarioService) Actualizar(ctx context.Context, id int, req *models.UsuarioUpdateRequest) (*models.UsuarioResponse, error) {
	normalizarUsuarioUpdate(req)
	if errs := validator.ValidateUsuarioUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	usuario, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	response := usuario.ToResponse()
	return &response, nil
}

func normalizarUsuarioCreate(req *models.UsuarioCreateRequest) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Apellido = strings.TrimSpace(req.Apellido)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Telefono = strings.TrimSpace(req.Telefono)
	req.Direccion = strings.TrimSpace(req.Direccion)
}

func normalizarUsuarioUpdate(req *models.UsuarioUpdateRequest) {
	trim := func(value **string, lower bool) {
		if *value == nil {
			return
		}
		clean := strings.TrimSpace(**value)
		if lower {
			clean = strings.ToLower(clean)
		}
		*value = &clean
	}
	trim(&req.Nombre, false)
	trim(&req.Apellido, false)
	trim(&req.Email, true)
	trim(&req.Telefono, false)
	trim(&req.Direccion, false)
}
