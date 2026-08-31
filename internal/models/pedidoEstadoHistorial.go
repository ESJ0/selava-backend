package models

import "time"

// PedidoEstadoHistorial representa un cambio de estado de un pedido y
// conserva el usuario responsable y el momento en que ocurrio.
type PedidoEstadoHistorial struct {
	ID            int       `json:"id"`
	PedidoID      int       `json:"pedido_id"`
	EstadoID      int       `json:"estado_id"`
	UsuarioID     int       `json:"usuario_id"`
	FechaCambio   time.Time `json:"fecha_cambio"`
	Observaciones *string   `json:"observaciones,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type PedidoEstadoUpdateRequest struct {
	EstadoID      int     `json:"estado_id"`
	Observaciones *string `json:"observaciones,omitempty"`
}
