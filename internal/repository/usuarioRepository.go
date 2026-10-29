package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")
	ErrEmailUsuarioEnUso   = errors.New("el email ya esta registrado")
)

const rolClienteID = 4

type UsuarioRepository struct {
	db *pgxpool.Pool
}

func NewUsuarioRepository(db *pgxpool.Pool) *UsuarioRepository { return &UsuarioRepository{db: db} }

type usuarioRow interface{ Scan(dest ...any) error }

func scanUsuario(row usuarioRow) (*models.Usuario, error) {
	var usuario models.Usuario
	err := row.Scan(
		&usuario.ID, &usuario.RolID, &usuario.Rol, &usuario.ClienteID,
		&usuario.Nombre, &usuario.Apellido, &usuario.Email, &usuario.PasswordHash,
		&usuario.Telefono, &usuario.Activo, &usuario.EmailVerificadoAt,
		&usuario.UltimoAccesoAt, &usuario.CreatedAt, &usuario.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsuarioNoEncontrado
		}
		return nil, err
	}
	return &usuario, nil
}

func (r *UsuarioRepository) GetByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	usuario, err := scanUsuario(r.db.QueryRow(ctx, `
		SELECT u.id, u.rol_id, r.nombre, u.cliente_id, u.nombre, u.apellido,
		       u.email, u.password_hash, u.telefono, u.activo,
		       u.email_verificado_at, u.ultimo_acceso_at, u.created_at, u.updated_at
		FROM usuarios u JOIN roles r ON r.id = u.rol_id
		WHERE u.email = LOWER(BTRIM($1))
	`, email))
	if err != nil {
		return nil, fmt.Errorf("obteniendo usuario por email: %w", err)
	}
	return usuario, nil
}

func (r *UsuarioRepository) GetByID(ctx context.Context, id int) (*models.Usuario, error) {
	usuario, err := scanUsuario(r.db.QueryRow(ctx, `
		SELECT u.id, u.rol_id, r.nombre, u.cliente_id, u.nombre, u.apellido,
		       u.email, u.password_hash, u.telefono, u.activo,
		       u.email_verificado_at, u.ultimo_acceso_at, u.created_at, u.updated_at
		FROM usuarios u JOIN roles r ON r.id = u.rol_id WHERE u.id = $1
	`, id))
	if err != nil {
		return nil, fmt.Errorf("obteniendo usuario: %w", err)
	}
	return usuario, nil
}

func (r *UsuarioRepository) List(ctx context.Context) ([]models.UsuarioResponse, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.rol_id, r.nombre, u.cliente_id, u.nombre, u.apellido,
		       u.email, u.telefono, u.activo, u.created_at, u.updated_at
		FROM usuarios u JOIN roles r ON r.id = u.rol_id
		ORDER BY u.created_at DESC, u.id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("listando usuarios: %w", err)
	}
	defer rows.Close()
	usuarios := make([]models.UsuarioResponse, 0)
	for rows.Next() {
		var usuario models.UsuarioResponse
		if err = rows.Scan(
			&usuario.ID, &usuario.RolID, &usuario.Rol, &usuario.ClienteID,
			&usuario.Nombre, &usuario.Apellido, &usuario.Email, &usuario.Telefono,
			&usuario.Activo, &usuario.CreatedAt, &usuario.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("leyendo usuario: %w", err)
		}
		usuarios = append(usuarios, usuario)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("recorriendo usuarios: %w", err)
	}
	return usuarios, nil
}

func (r *UsuarioRepository) Create(ctx context.Context, req *models.UsuarioCreateRequest, passwordHash string) (*models.Usuario, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciando creacion de usuario: %w", err)
	}
	defer tx.Rollback(ctx)
	var clienteID *int
	if req.RolID == rolClienteID {
		var id int
		err = tx.QueryRow(ctx, `
			INSERT INTO clientes(nombre, apellido, telefono, email, direccion)
			VALUES($1, $2, $3, LOWER(BTRIM($4)), $5) RETURNING id
		`, req.Nombre, req.Apellido, req.Telefono, req.Email, req.Direccion).Scan(&id)
		if err != nil {
			return nil, usuarioWriteError("creando perfil de cliente", err)
		}
		clienteID = &id
	}
	var usuarioID int
	err = tx.QueryRow(ctx, `
		INSERT INTO usuarios(rol_id, cliente_id, nombre, apellido, email, password_hash, telefono)
		VALUES($1, $2, $3, $4, LOWER(BTRIM($5)), $6, $7) RETURNING id
	`, req.RolID, clienteID, req.Nombre, req.Apellido, req.Email, passwordHash, req.Telefono).Scan(&usuarioID)
	if err != nil {
		return nil, usuarioWriteError("creando usuario", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmando creacion de usuario: %w", err)
	}
	return r.GetByID(ctx, usuarioID)
}

func (r *UsuarioRepository) Update(ctx context.Context, id int, req *models.UsuarioUpdateRequest) (*models.Usuario, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciando actualizacion de usuario: %w", err)
	}
	defer tx.Rollback(ctx)
	var clienteID *int
	err = tx.QueryRow(ctx, `
		UPDATE usuarios SET
			nombre = COALESCE($2, nombre), apellido = COALESCE($3, apellido),
			email = COALESCE(LOWER(BTRIM($4)), email), telefono = COALESCE($5, telefono),
			activo = COALESCE($6, activo), updated_at = NOW()
		WHERE id = $1 RETURNING cliente_id
	`, id, req.Nombre, req.Apellido, req.Email, req.Telefono, req.Activo).Scan(&clienteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsuarioNoEncontrado
		}
		return nil, usuarioWriteError("actualizando usuario", err)
	}
	if clienteID != nil {
		_, err = tx.Exec(ctx, `
			UPDATE clientes SET nombre = COALESCE($2, nombre), apellido = COALESCE($3, apellido),
				email = COALESCE(LOWER(BTRIM($4)), email), telefono = COALESCE($5, telefono),
				direccion = COALESCE($6, direccion), updated_at = NOW()
			WHERE id = $1
		`, *clienteID, req.Nombre, req.Apellido, req.Email, req.Telefono, req.Direccion)
		if err != nil {
			return nil, usuarioWriteError("actualizando perfil de cliente", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmando actualizacion de usuario: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *UsuarioRepository) MarcarUltimoAcceso(ctx context.Context, usuarioID int) error {
	tag, err := r.db.Exec(ctx, `UPDATE usuarios SET ultimo_acceso_at = NOW(), updated_at = NOW() WHERE id = $1`, usuarioID)
	if err != nil {
		return fmt.Errorf("actualizando ultimo acceso: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUsuarioNoEncontrado
	}
	return nil
}

func usuarioWriteError(operacion string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrEmailUsuarioEnUso
	}
	return fmt.Errorf("%s: %w", operacion, err)
}
