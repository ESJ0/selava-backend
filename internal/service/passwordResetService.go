package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

const passwordResetTTL = 30 * time.Minute

type PasswordResetService struct {
	usuarios *repository.UsuarioRepository
	tokens   *repository.PasswordResetRepository
	mailer   PasswordMailer
	resetURL string
}

func NewPasswordResetService(usuarios *repository.UsuarioRepository, tokens *repository.PasswordResetRepository, mailer PasswordMailer, resetURL string) *PasswordResetService {
	return &PasswordResetService{usuarios: usuarios, tokens: tokens, mailer: mailer, resetURL: strings.TrimRight(resetURL, "?")}
}

func (s *PasswordResetService) Solicitar(ctx context.Context, email string) error {
	usuario, err := s.usuarios.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, repository.ErrUsuarioNoEncontrado) {
			return nil
		}
		return err
	}
	if !usuario.Activo {
		return nil
	}
	raw, err := tokenSeguro()
	if err != nil {
		return err
	}
	if err = s.tokens.Crear(ctx, usuario.ID, hashToken(raw), time.Now().Add(passwordResetTTL)); err != nil {
		return err
	}
	separator := "?"
	if strings.Contains(s.resetURL, "?") {
		separator = "&"
	}
	link := s.resetURL + separator + "token=" + url.QueryEscape(raw)
	return s.mailer.EnviarRecuperacion(ctx, usuario.Email, usuario.Nombre, link)
}

func (s *PasswordResetService) Validar(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return repository.ErrTokenRecuperacionInvalido
	}
	return s.tokens.Validar(ctx, hashToken(token))
}

func (s *PasswordResetService) Restablecer(ctx context.Context, token, password string) error {
	if errs := validator.ValidatePassword(password); errs.HasErrors() {
		return errs
	}
	if strings.TrimSpace(token) == "" {
		return repository.ErrTokenRecuperacionInvalido
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("protegiendo nueva password: %w", err)
	}
	return s.tokens.Consumir(ctx, hashToken(token), hash)
}

func tokenSeguro() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generando token seguro: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
