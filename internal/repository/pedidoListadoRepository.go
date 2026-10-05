package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
)

func (r *PedidoRepository) List(ctx context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error) {
	conditions := []string{"TRUE"}
	args := []any{}
	bind := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}
	if filter.Buscar != "" {
		literal := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(filter.Buscar)
		pattern := bind("%" + literal + "%")
		search := fmt.Sprintf("(concat_ws(' ', c.nombre, c.apellido) ILIKE %s OR c.telefono ILIKE %s", pattern, pattern)
		identifier := strings.TrimPrefix(strings.ToUpper(filter.Buscar), "#")
		identifier = strings.TrimPrefix(identifier, "SLV-")
		if id, err := strconv.Atoi(identifier); err == nil && id > 0 && id <= 2147483647 {
			search += " OR p.id = " + bind(id)
		}
		conditions = append(conditions, search+")")
	}
	if filter.EstadoID != 0 {
		conditions = append(conditions, "p.estado_actual_id = "+bind(filter.EstadoID))
	}
	if filter.FechaDesde != nil {
		conditions = append(conditions, "p.fecha_recibido >= "+bind(*filter.FechaDesde))
	}
	if filter.FechaHastaExclusiva != nil {
		conditions = append(conditions, "p.fecha_recibido < "+bind(*filter.FechaHastaExclusiva))
	}
	from := ` FROM pedidos p JOIN clientes c ON c.id = p.cliente_id
		JOIN estados_pedido e ON e.id = p.estado_actual_id WHERE ` + strings.Join(conditions, " AND ")
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("error iniciando listado de pedidos: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	result := &models.PedidoListResponse{Pedidos: []models.PedidoResumen{}, Pagina: filter.Pagina, Limite: filter.Limite}
	if err := tx.QueryRow(ctx, "SELECT COUNT(*)"+from, args...).Scan(&result.Total); err != nil {
		return nil, fmt.Errorf("error contando pedidos: %w", err)
	}
	direction := "DESC"
	if filter.Orden == "antiguos" {
		direction = "ASC"
	}
	limit := bind(filter.Limite)
	offset := bind((filter.Pagina - 1) * filter.Limite)
	query := `SELECT p.id, p.fecha_recibido, p.fecha_entrega_estimada, p.total, p.activo,
		c.id, c.nombre, c.apellido, c.telefono, e.id, e.nombre, e.orden` + from +
		" ORDER BY p.fecha_recibido " + direction + ", p.id " + direction + " LIMIT " + limit + " OFFSET " + offset
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando pedidos: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item models.PedidoResumen
		if err := rows.Scan(&item.ID, &item.FechaRecibido, &item.FechaEntregaEstimada, &item.Total, &item.Activo,
			&item.Cliente.ID, &item.Cliente.Nombre, &item.Cliente.Apellido, &item.Cliente.Telefono,
			&item.EstadoActual.ID, &item.EstadoActual.Nombre, &item.EstadoActual.Orden); err != nil {
			return nil, fmt.Errorf("error leyendo pedido: %w", err)
		}
		result.Pedidos = append(result.Pedidos, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo pedidos: %w", err)
	}
	rows.Close()
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error completando listado de pedidos: %w", err)
	}
	return result, nil
}
