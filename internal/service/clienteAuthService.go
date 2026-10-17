package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

var ErrCredencialesClienteInvalidas = errors.New("email o contrasena incorrectos")

type cuentaClienteRepository interface {
	Registrar(context.Context, *models.ClienteRegistroRequest, string) (*models.ClienteRegistrado, error)
	GetByEmail(context.Context, string) (*models.ClienteRegistrado, error)
	GetClienteActivoByID(context.Context, int) (*models.Cliente, error)
	MarcarUltimoAcceso(context.Context, int) error
}

type ClienteAuthService struct {
	repo      cuentaClienteRepository
	jwtSecret string
}

func NewClienteAuthService(repo cuentaClienteRepository, jwtSecret string) *ClienteAuthService {
	return &ClienteAuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *ClienteAuthService) Registrar(ctx context.Context, req *models.ClienteRegistroRequest) (*models.ClienteRegistrado, error) {
	normalizarRegistroCliente(req)
	if errs := validator.ValidateClienteRegistro(req); errs.HasErrors() {
		return nil, errs
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("protegiendo password de cliente: %w", err)
	}
	return s.repo.Registrar(ctx, req, passwordHash)
}

func (s *ClienteAuthService) Login(ctx context.Context, req *models.ClienteLoginRequest) (string, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if errs := validator.ValidateClienteLogin(req); errs.HasErrors() {
		return "", errs
	}

	registro, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrCuentaClienteNoEncontrada) {
			return "", ErrCredencialesClienteInvalidas
		}
		return "", err
	}
	if !registro.Cliente.Activo || !registro.Cuenta.Activo || !auth.CheckPassword(registro.Cuenta.PasswordHash, req.Password) {
		return "", ErrCredencialesClienteInvalidas
	}

	token, err := auth.GenerateClienteToken(s.jwtSecret, registro.Cliente.ID)
	if err != nil {
		return "", fmt.Errorf("generando token de cliente: %w", err)
	}
	if err = s.repo.MarcarUltimoAcceso(ctx, registro.Cuenta.ID); err != nil {
		return "", err
	}
	return token, nil
}

func (s *ClienteAuthService) ObtenerPerfil(ctx context.Context, clienteID int) (*models.ClienteResponse, error) {
	if clienteID <= 0 {
		return nil, ErrCredencialesClienteInvalidas
	}
	cliente, err := s.repo.GetClienteActivoByID(ctx, clienteID)
	if err != nil {
		if errors.Is(err, repository.ErrCuentaClienteNoEncontrada) {
			return nil, ErrCredencialesClienteInvalidas
		}
		return nil, err
	}
	response := cliente.ToResponse()
	return &response, nil
}

func normalizarRegistroCliente(req *models.ClienteRegistroRequest) {
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Apellido = strings.TrimSpace(req.Apellido)
	req.Telefono = strings.TrimSpace(req.Telefono)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Direccion = strings.TrimSpace(req.Direccion)
}
