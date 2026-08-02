package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
)

type fakeUsuarioRepo struct {
	usuario *models.Usuario
	err     error
	email   string
}

func (r *fakeUsuarioRepo) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	r.email = email
	if r.err != nil {
		return nil, r.err
	}
	return r.usuario, nil
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	hash, err := auth.HashPassword("clave-segura")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	repo := &fakeUsuarioRepo{usuario: &models.Usuario{
		ID:           10,
		RolID:        1,
		Email:        "admin@selava.com",
		PasswordHash: hash,
	}}
	service := NewAuthService(repo, "secret")

	token, err := service.Login(context.Background(), "admin@selava.com", "clave-segura")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if repo.email != "admin@selava.com" {
		t.Fatalf("expected repo email admin@selava.com, got %q", repo.email)
	}

	claims, err := auth.ValidateToken("secret", token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims.UsuarioID != 10 || claims.RolID != 1 {
		t.Fatalf("unexpected claims: usuario_id=%d rol_id=%d", claims.UsuarioID, claims.RolID)
	}
}

func TestAuthServiceLoginRejectsUnknownUser(t *testing.T) {
	service := NewAuthService(&fakeUsuarioRepo{err: repository.ErrUsuarioNoEncontrado}, "secret")

	_, err := service.Login(context.Background(), "nadie@selava.com", "clave")
	if !errors.Is(err, ErrCredencialesInvalidas) {
		t.Fatalf("expected ErrCredencialesInvalidas, got %v", err)
	}
}

func TestAuthServiceLoginRejectsWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("clave-segura")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	service := NewAuthService(&fakeUsuarioRepo{usuario: &models.Usuario{PasswordHash: hash}}, "secret")

	_, err = service.Login(context.Background(), "admin@selava.com", "otra-clave")
	if !errors.Is(err, ErrCredencialesInvalidas) {
		t.Fatalf("expected ErrCredencialesInvalidas, got %v", err)
	}
}

func TestAuthServiceLoginPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	service := NewAuthService(&fakeUsuarioRepo{err: repoErr}, "secret")

	_, err := service.Login(context.Background(), "admin@selava.com", "clave")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
