package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClienteNoEncontrado = errors.New("cliente no encontrado")

type ClienteRepository struct {
	db *pgxpool.Pool
}

func NewClienteRepository(db *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{db: db}
}

func (r *ClienteRepository) Create(ctx context.Context, c *models.ClienteCreateRequest) (*models.Cliente, error) {
	query := `
		INSERT INTO clientes (nombre, apellido, telefono, email, direccion)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, nombre, apellido, telefono, email, direccion, activo, created_at, updated_at
	`
	var cliente models.Cliente
	err := r.db.QueryRow(ctx, query, c.Nombre, c.Apellido, c.Telefono, c.Email, c.Direccion).Scan(
		&cliente.ID, &cliente.Nombre, &cliente.Apellido, &cliente.Telefono,
		&cliente.Email, &cliente.Direccion, &cliente.Activo, &cliente.CreatedAt, &cliente.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %w", err)
	}
	return &cliente, nil
}

func (r *ClienteRepository) GetByID(ctx context.Context, id int) (*models.Cliente, error) {
	query := `
		SELECT id, nombre, apellido, telefono, email, direccion, activo, created_at, updated_at
		FROM clientes
		WHERE id = $1
	`
	var cliente models.Cliente
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cliente.ID, &cliente.Nombre, &cliente.Apellido, &cliente.Telefono,
		&cliente.Email, &cliente.Direccion, &cliente.Activo, &cliente.CreatedAt, &cliente.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClienteNoEncontrado
		}
		return nil, fmt.Errorf("error obteniendo cliente: %w", err)
	}
	return &cliente, nil
}

func (r *ClienteRepository) List(ctx context.Context) ([]models.Cliente, error) {
	query := `
		SELECT id, nombre, apellido, telefono, email, direccion, activo, created_at, updated_at
		FROM clientes
		ORDER BY id
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando clientes: %w", err)
	}
	defer rows.Close()

	var clientes []models.Cliente
	for rows.Next() {
		var cliente models.Cliente
		if err := rows.Scan(
			&cliente.ID, &cliente.Nombre, &cliente.Apellido, &cliente.Telefono,
			&cliente.Email, &cliente.Direccion, &cliente.Activo, &cliente.CreatedAt, &cliente.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error leyendo cliente: %w", err)
		}
		clientes = append(clientes, cliente)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando clientes: %w", err)
	}
	return clientes, nil
}

func (r *ClienteRepository) Update(ctx context.Context, id int, u *models.ClienteUpdateRequest) (*models.Cliente, error) {
	// COALESCE permite actualizar solo los campos que vienen en el request
	query := `
		UPDATE clientes SET
			nombre     = COALESCE($1, nombre),
			apellido   = COALESCE($2, apellido),
			telefono   = COALESCE($3, telefono),
			email      = COALESCE($4, email),
			direccion  = COALESCE($5, direccion),
			activo     = COALESCE($6, activo),
			updated_at = NOW()
		WHERE id = $7
		RETURNING id, nombre, apellido, telefono, email, direccion, activo, created_at, updated_at
	`
	var cliente models.Cliente
	err := r.db.QueryRow(ctx, query,
		u.Nombre, u.Apellido, u.Telefono, u.Email, u.Direccion, u.Activo, id,
	).Scan(
		&cliente.ID, &cliente.Nombre, &cliente.Apellido, &cliente.Telefono,
		&cliente.Email, &cliente.Direccion, &cliente.Activo, &cliente.CreatedAt, &cliente.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClienteNoEncontrado
		}
		return nil, fmt.Errorf("error actualizando cliente: %w", err)
	}
	return &cliente, nil
}

func (r *ClienteRepository) Delete(ctx context.Context, id int) error {
	// Borrado lógico: se marca como inactivo en lugar de eliminar el registro
	query := `UPDATE clientes SET activo = FALSE, updated_at = NOW() WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error eliminando cliente: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrClienteNoEncontrado
	}
	return nil
}

func (r *ClienteRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM clientes WHERE email = $1)`
	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("error verificando email: %w", err)
	}
	return exists, nil
}
