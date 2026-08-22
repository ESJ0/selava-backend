package models

import "time"

// Pedido representa un registro de la tabla pedidos.
type Pedido struct {
	ID                   int        `json:"id"`
	ClienteID            int        `json:"cliente_id"`
	UsuarioID            int        `json:"usuario_id"`
	EstadoActualID       int        `json:"estado_actual_id"`
	FechaRecibido        time.Time  `json:"fecha_recibido"`
	FechaEntregaEstimada *time.Time `json:"fecha_entrega_estimada,omitempty"`
	FechaEntregaReal     *time.Time `json:"fecha_entrega_real,omitempty"`
	Total                float64    `json:"total"`
	Observaciones        *string    `json:"observaciones,omitempty"`
	Activo               bool       `json:"activo"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
