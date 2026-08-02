package service

import (
	"context"
	"errors"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
)

var ErrCredencialesInvalidas = errors.New("email o contrasena incorrectos")

type UsuarioRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.Usuario, error)
}

type AuthService struct {
	usuarioRepo UsuarioRepository
	jwtSecret   string
}

func NewAuthService(usuarioRepo UsuarioRepository, jwtSecret string) *AuthService {
	return &AuthService{usuarioRepo: usuarioRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	usuario, err := s.usuarioRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUsuarioNoEncontrado) {
			return "", ErrCredencialesInvalidas
		}
		return "", err
	}

	if !auth.CheckPassword(usuario.PasswordHash, password) {
		return "", ErrCredencialesInvalidas
	}

	return auth.GenerateToken(s.jwtSecret, usuario.ID, usuario.RolID)
}
