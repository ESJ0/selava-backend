package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTokenRecuperacionInvalido = errors.New("token de recuperacion invalido o vencido")

type PasswordResetRepository struct{ db *pgxpool.Pool }

func NewPasswordResetRepository(db *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Crear(ctx context.Context, usuarioID int, tokenHash string, expiresAt time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciando token de recuperacion: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE usuario_id = $1 AND used_at IS NULL`, usuarioID); err != nil {
		return fmt.Errorf("invalidando tokens anteriores: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO password_reset_tokens(usuario_id, token_hash, expires_at) VALUES($1,$2,$3)`, usuarioID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("creando token de recuperacion: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *PasswordResetRepository) Validar(ctx context.Context, tokenHash string) error {
	var existe bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM password_reset_tokens t
			JOIN usuarios u ON u.id = t.usuario_id
			WHERE t.token_hash = $1 AND t.used_at IS NULL AND t.expires_at > NOW() AND u.activo = TRUE
		)
	`, tokenHash).Scan(&existe)
	if err != nil {
		return fmt.Errorf("validando token de recuperacion: %w", err)
	}
	if !existe {
		return ErrTokenRecuperacionInvalido
	}
	return nil
}

func (r *PasswordResetRepository) Consumir(ctx context.Context, tokenHash, passwordHash string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciando cambio de password: %w", err)
	}
	defer tx.Rollback(ctx)
	var tokenID, usuarioID int
	err = tx.QueryRow(ctx, `
		SELECT t.id, t.usuario_id FROM password_reset_tokens t
		JOIN usuarios u ON u.id = t.usuario_id
		WHERE t.token_hash = $1 AND t.used_at IS NULL AND t.expires_at > NOW() AND u.activo = TRUE
		FOR UPDATE OF t
	`, tokenHash).Scan(&tokenID, &usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTokenRecuperacionInvalido
		}
		return fmt.Errorf("leyendo token de recuperacion: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE usuarios SET password_hash=$2, updated_at=NOW() WHERE id=$1`, usuarioID, passwordHash); err != nil {
		return fmt.Errorf("actualizando password: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at=NOW() WHERE id=$1`, tokenID); err != nil {
		return fmt.Errorf("consumiendo token: %w", err)
	}
	return tx.Commit(ctx)
}
