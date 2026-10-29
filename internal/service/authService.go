package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
)

var ErrCredencialesInvalidas = errors.New("email o contraseña incorrectos")

type AuthService struct {
	usuarioRepo *repository.UsuarioRepository
	jwtSecret   string
}

func NewAuthService(usuarioRepo *repository.UsuarioRepository, jwtSecret string) *AuthService {
	return &AuthService{usuarioRepo: usuarioRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	resultado, err := s.LoginUnificado(ctx, email, password)
	if err != nil {
		return "", err
	}
	return resultado.Token, nil
}

func (s *AuthService) LoginUnificado(ctx context.Context, email, password string) (*models.LoginUnificadoResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	usuario, err := s.usuarioRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUsuarioNoEncontrado) {
			return nil, ErrCredencialesInvalidas
		}
		return nil, err
	}

	if !usuario.Activo || !auth.CheckPassword(usuario.PasswordHash, password) {
		return nil, ErrCredencialesInvalidas
	}

	var token, destino string
	if usuario.RolID == 4 && usuario.ClienteID != nil {
		token, err = auth.GenerateClienteAccountToken(s.jwtSecret, usuario.ID, *usuario.ClienteID)
		destino = "cliente"
	} else {
		token, err = auth.GenerateTokenConRol(s.jwtSecret, usuario.ID, usuario.RolID, usuario.Rol)
		destino = "admin"
	}
	if err != nil {
		return nil, err
	}
	if err = s.usuarioRepo.MarcarUltimoAcceso(ctx, usuario.ID); err != nil {
		return nil, err
	}
	return &models.LoginUnificadoResponse{Token: token, Usuario: usuario.ToResponse(), Destino: destino}, nil
}
