package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")

type UsuarioRepository struct {
	db *pgxpool.Pool
}

func NewUsuarioRepository(db *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

func (r *UsuarioRepository) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	query := `
		SELECT id, rol_id, nombre, apellido, email, password_hash, telefono, activo, created_at, updated_at
		FROM usuarios
		WHERE email = $1 AND activo = TRUE
	`
	var u models.Usuario
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.RolID, &u.Nombre, &u.Apellido, &u.Email, &u.PasswordHash,
		&u.Telefono, &u.Activo, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsuarioNoEncontrado
		}
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}
	return &u, nil
}
