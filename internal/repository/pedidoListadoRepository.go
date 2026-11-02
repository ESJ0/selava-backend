package repository

import (
	"context"
	"fmt"
	"math"
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
	if filter.ClienteID > 0 {
		conditions = append(conditions, "p.cliente_id = "+bind(filter.ClienteID))
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
	if filter.SoloVigentes {
		conditions = append(conditions, "p.activo = TRUE")
	}
	switch filter.Seccion {
	case "activos":
		conditions = append(conditions, "NOT e.es_terminal")
	case "historial":
		conditions = append(conditions, "e.es_terminal")
	}
	if filter.TipoPrendaID != 0 {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM prendas pr WHERE pr.pedido_id = p.id AND pr.tipo_prenda_id = "+bind(filter.TipoPrendaID)+")")
	}
	switch filter.Saldo {
	case "pendiente":
		conditions = append(conditions, "NOT e.es_cancelacion AND p.total - pg.pagado > 0.004")
	case "pagado":
		conditions = append(conditions, "p.total - pg.pagado <= 0.004")
	}
	from := ` FROM pedidos p JOIN clientes c ON c.id = p.cliente_id
		JOIN estados_pedido e ON e.id = p.estado_actual_id
		LEFT JOIN LATERAL (SELECT COALESCE(SUM(pago.monto), 0)::float8 AS pagado FROM pagos pago WHERE pago.pedido_id = p.id) pg ON TRUE
		WHERE ` + strings.Join(conditions, " AND ")
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
	orderBy := "p.fecha_recibido " + direction + ", p.id " + direction
	if filter.Orden == "prioridad" {
		// Primero los pedidos en proceso y luego los terminados, cada grupo del más reciente al más antiguo.
		orderBy = "e.es_terminal ASC, p.fecha_recibido DESC, p.id DESC"
	}
	limit := bind(filter.Limite)
	offset := bind((filter.Pagina - 1) * filter.Limite)
	query := `SELECT p.id, p.fecha_recibido, p.fecha_entrega_estimada, p.total, p.activo,
		c.id, c.nombre, c.apellido, c.telefono, e.id, e.nombre, e.orden,
		pg.pagado, NOT e.es_terminal, e.es_cancelacion,
		ARRAY(SELECT DISTINCT t.nombre FROM prendas pr JOIN tipos_prenda t ON t.id = pr.tipo_prenda_id WHERE pr.pedido_id = p.id ORDER BY t.nombre)::text[],
		COALESCE((SELECT SUM(pr.cantidad) FROM prendas pr WHERE pr.pedido_id = p.id), 0)::int` + from +
		" ORDER BY " + orderBy + " LIMIT " + limit + " OFFSET " + offset
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando pedidos: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item models.PedidoResumen
		if err := rows.Scan(&item.ID, &item.FechaRecibido, &item.FechaEntregaEstimada, &item.Total, &item.Activo,
			&item.Cliente.ID, &item.Cliente.Nombre, &item.Cliente.Apellido, &item.Cliente.Telefono,
			&item.EstadoActual.ID, &item.EstadoActual.Nombre, &item.EstadoActual.Orden,
			&item.TotalPagado, &item.EnProceso, &item.Cancelado, &item.TiposPrenda, &item.CantidadPrendas); err != nil {
			return nil, fmt.Errorf("error leyendo pedido: %w", err)
		}
		// Mismo cálculo oficial que /saldo: centavos redondeados y saldo nunca negativo.
		item.Total = math.Round(item.Total*100) / 100
		item.TotalPagado = math.Round(item.TotalPagado*100) / 100
		item.SaldoPendiente = math.Max(0, math.Round((item.Total-item.TotalPagado)*100)/100)
		if item.TiposPrenda == nil {
			item.TiposPrenda = []string{}
		}
		result.Pedidos = append(result.Pedidos, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo pedidos: %w", err)
	}
	rows.Close()
	if filter.ClienteID > 0 {
		totales := &models.PedidosClienteTotales{}
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FILTER (WHERE NOT e.es_terminal),
			       COUNT(*) FILTER (WHERE e.es_terminal),
			       COUNT(*) FILTER (WHERE NOT e.es_cancelacion AND p.total - pg.pagado > 0.004),
			       COALESCE(SUM(GREATEST(p.total - pg.pagado, 0)) FILTER (WHERE NOT e.es_cancelacion), 0)::float8,
			       COALESCE(SUM(pg.pagado), 0)::float8,
			       COALESCE(SUM(p.total) FILTER (WHERE NOT e.es_cancelacion), 0)::float8
			FROM pedidos p
			JOIN estados_pedido e ON e.id = p.estado_actual_id
			LEFT JOIN LATERAL (SELECT COALESCE(SUM(pago.monto), 0)::float8 AS pagado FROM pagos pago WHERE pago.pedido_id = p.id) pg ON TRUE
			WHERE p.cliente_id = $1 AND p.activo = TRUE`, filter.ClienteID).Scan(
			&totales.Activos, &totales.Historial, &totales.PedidosConSaldo, &totales.SaldoPendiente, &totales.TotalPagado, &totales.TotalFacturado)
		if err != nil {
			return nil, fmt.Errorf("error resumiendo pedidos del cliente: %w", err)
		}
		totales.SaldoPendiente = math.Round(totales.SaldoPendiente*100) / 100
		totales.TotalPagado = math.Round(totales.TotalPagado*100) / 100
		totales.TotalFacturado = math.Round(totales.TotalFacturado*100) / 100
		result.Resumen = totales
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error completando listado de pedidos: %w", err)
	}
	return result, nil
}

// TiposPrendaCliente lista los tipos de prenda que aparecen en los pedidos vigentes del cliente.
func (r *PedidoRepository) TiposPrendaCliente(ctx context.Context, clienteID int) ([]models.TipoPrendaOpcion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT t.id, t.nombre
		FROM pedidos p
		JOIN prendas pr ON pr.pedido_id = p.id
		JOIN tipos_prenda t ON t.id = pr.tipo_prenda_id
		WHERE p.cliente_id = $1 AND p.activo = TRUE
		ORDER BY t.nombre`, clienteID)
	if err != nil {
		return nil, fmt.Errorf("error consultando tipos de prenda del cliente: %w", err)
	}
	defer rows.Close()
	tipos := []models.TipoPrendaOpcion{}
	for rows.Next() {
		var tipo models.TipoPrendaOpcion
		if err := rows.Scan(&tipo.ID, &tipo.Nombre); err != nil {
			return nil, fmt.Errorf("error leyendo tipo de prenda: %w", err)
		}
		tipos = append(tipos, tipo)
	}
	return tipos, rows.Err()
}
