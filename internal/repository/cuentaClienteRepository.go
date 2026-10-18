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
	ErrCuentaClienteNoEncontrada = errors.New("cuenta de cliente no encontrada")
	ErrEmailCuentaClienteEnUso   = errors.New("el email ya esta registrado")
)

type CuentaClienteRepository struct {
	db *pgxpool.Pool
}

func NewCuentaClienteRepository(db *pgxpool.Pool) *CuentaClienteRepository {
	return &CuentaClienteRepository{db: db}
}

// Registrar crea el perfil y su cuenta como una sola operacion. Si cualquiera
// de los INSERT falla, ninguna de las dos filas queda persistida.
func (r *CuentaClienteRepository) Registrar(ctx context.Context, req *models.ClienteRegistroRequest, passwordHash string) (*models.ClienteRegistrado, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciando registro de cliente: %w", err)
	}
	defer tx.Rollback(ctx)

	var resultado models.ClienteRegistrado
	err = tx.QueryRow(ctx, `
		INSERT INTO clientes(nombre, apellido, telefono, email, direccion)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, nombre, apellido, telefono, email, direccion, activo, created_at, updated_at
	`, req.Nombre, req.Apellido, req.Telefono, req.Email, req.Direccion).Scan(
		&resultado.Cliente.ID,
		&resultado.Cliente.Nombre,
		&resultado.Cliente.Apellido,
		&resultado.Cliente.Telefono,
		&resultado.Cliente.Email,
		&resultado.Cliente.Direccion,
		&resultado.Cliente.Activo,
		&resultado.Cliente.CreatedAt,
		&resultado.Cliente.UpdatedAt,
	)
	if err != nil {
		return nil, cuentaClienteWriteError("creando perfil de cliente", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO cuentas_cliente(cliente_id, email, password_hash)
		VALUES($1, $2, $3)
		RETURNING id, cliente_id, email, password_hash, activo,
		          email_verificado_at, ultimo_acceso_at, created_at, updated_at
	`, resultado.Cliente.ID, req.Email, passwordHash).Scan(
		&resultado.Cuenta.ID,
		&resultado.Cuenta.ClienteID,
		&resultado.Cuenta.Email,
		&resultado.Cuenta.PasswordHash,
		&resultado.Cuenta.Activo,
		&resultado.Cuenta.EmailVerificadoAt,
		&resultado.Cuenta.UltimoAccesoAt,
		&resultado.Cuenta.CreatedAt,
		&resultado.Cuenta.UpdatedAt,
	)
	if err != nil {
		return nil, cuentaClienteWriteError("creando cuenta de cliente", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmando registro de cliente: %w", err)
	}
	return &resultado, nil
}

// GetByEmail devuelve tambien el estado del perfil para que el servicio exija
// que tanto el cliente como su cuenta sigan activos.
func (r *CuentaClienteRepository) GetByEmail(ctx context.Context, email string) (*models.ClienteRegistrado, error) {
	var resultado models.ClienteRegistrado
	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.nombre, c.apellido, c.telefono,
		       COALESCE(c.email, ''), COALESCE(c.direccion, ''),
		       c.activo, c.created_at, c.updated_at,
		       cc.id, cc.cliente_id, cc.email, cc.password_hash, cc.activo,
		       cc.email_verificado_at, cc.ultimo_acceso_at, cc.created_at, cc.updated_at
		FROM cuentas_cliente cc
		JOIN clientes c ON c.id = cc.cliente_id
		WHERE cc.email = $1
	`, email).Scan(
		&resultado.Cliente.ID,
		&resultado.Cliente.Nombre,
		&resultado.Cliente.Apellido,
		&resultado.Cliente.Telefono,
		&resultado.Cliente.Email,
		&resultado.Cliente.Direccion,
		&resultado.Cliente.Activo,
		&resultado.Cliente.CreatedAt,
		&resultado.Cliente.UpdatedAt,
		&resultado.Cuenta.ID,
		&resultado.Cuenta.ClienteID,
		&resultado.Cuenta.Email,
		&resultado.Cuenta.PasswordHash,
		&resultado.Cuenta.Activo,
		&resultado.Cuenta.EmailVerificadoAt,
		&resultado.Cuenta.UltimoAccesoAt,
		&resultado.Cuenta.CreatedAt,
		&resultado.Cuenta.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCuentaClienteNoEncontrada
		}
		return nil, fmt.Errorf("obteniendo cuenta de cliente: %w", err)
	}
	return &resultado, nil
}

// GetClienteActivoByID carga el perfil asociado a una cuenta activa. Se usa
// para restaurar sesiones sin confiar en datos de perfil almacenados por el
// navegador.
func (r *CuentaClienteRepository) GetClienteActivoByID(ctx context.Context, clienteID int) (*models.Cliente, error) {
	var cliente models.Cliente
	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.nombre, c.apellido, c.telefono,
		       COALESCE(c.email, ''), COALESCE(c.direccion, ''),
		       c.activo, c.created_at, c.updated_at
		FROM cuentas_cliente cc
		JOIN clientes c ON c.id = cc.cliente_id
		WHERE cc.cliente_id = $1 AND cc.activo = TRUE AND c.activo = TRUE
	`, clienteID).Scan(
		&cliente.ID, &cliente.Nombre, &cliente.Apellido, &cliente.Telefono,
		&cliente.Email, &cliente.Direccion, &cliente.Activo,
		&cliente.CreatedAt, &cliente.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCuentaClienteNoEncontrada
		}
		return nil, fmt.Errorf("obteniendo perfil de cuenta cliente: %w", err)
	}
	return &cliente, nil
}

func (r *CuentaClienteRepository) MarcarUltimoAcceso(ctx context.Context, cuentaID int) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE cuentas_cliente
		SET ultimo_acceso_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, cuentaID)
	if err != nil {
		return fmt.Errorf("actualizando ultimo acceso: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCuentaClienteNoEncontrada
	}
	return nil
}

func cuentaClienteWriteError(operacion string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrEmailCuentaClienteEnUso
	}
	return fmt.Errorf("%s: %w", operacion, err)
}
