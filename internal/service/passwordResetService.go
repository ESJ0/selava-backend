package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

const (
	passwordResetTTL = 30 * time.Minute
	mailTimeout      = 45 * time.Second
)

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
	email = strings.ToLower(strings.TrimSpace(email))
	usuario, err := s.usuarios.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUsuarioNoEncontrado) {
			// Solo en el log del servidor: al cliente siempre se le responde lo mismo.
			log.Printf("recuperacion de contrasena: no existe una cuenta con el correo %s; no se envia correo", email)
			return nil
		}
		return err
	}
	if !usuario.Activo {
		log.Printf("recuperacion de contrasena: la cuenta %s esta inactiva; no se envia correo", email)
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
	// El envío se hace fuera de la petición HTTP (que tiene pocos segundos de límite) para que
	// un servidor SMTP lento no corte el correo; los errores quedan en el log.
	send := func() {
		mailCtx, cancel := context.WithTimeout(context.Background(), mailTimeout)
		defer cancel()
		if err := s.mailer.EnviarRecuperacion(mailCtx, usuario.Email, usuario.Nombre, link); err != nil {
			log.Printf("no se pudo enviar el correo de recuperacion a %s: %v", usuario.Email, err)
			return
		}
		log.Printf("correo de recuperacion enviado a %s", usuario.Email)
	}
	go send()
	return nil
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
