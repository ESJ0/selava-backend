package models

import "time"

// PedidoPropioDetalle es la vista segura del detalle para el cliente dueño.
// No expone IDs ni datos personales de usuarios internos.
type PedidoPropioDetalle struct {
	ID                   int          `json:"id"`
	FechaRecibido        time.Time    `json:"fecha_recibido"`
	FechaEntregaEstimada *time.Time   `json:"fecha_entrega_estimada,omitempty"`
	FechaEntregaReal     *time.Time   `json:"fecha_entrega_real,omitempty"`
	Total                float64      `json:"total"`
	Observaciones        *string      `json:"observaciones,omitempty"`
	Prendas              []Prenda     `json:"prendas"`
	EstadoActual         EstadoPedido `json:"estado_actual"`
}

// PedidoPropioHistorial representa un movimiento visible en el timeline del
// cliente sin revelar al empleado responsable del cambio.
type PedidoPropioHistorial struct {
	ID            int          `json:"id"`
	PedidoID      int          `json:"pedido_id"`
	Estado        EstadoPedido `json:"estado"`
	FechaCambio   time.Time    `json:"fecha_cambio"`
	Observaciones *string      `json:"observaciones,omitempty"`
}
