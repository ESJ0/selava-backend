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

// UsuarioResumen es la version minima de un usuario que se expone dentro
// de otras respuestas (como el historial de estados), sin datos sensibles.
type UsuarioResumen struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
}

// PedidoEstadoHistorialDetalle es una entrada del historial con el estado
// y el usuario responsable ya resueltos a nombre, para que el frontend no
// tenga que cruzar estado_id/usuario_id contra otros catalogos.
type PedidoEstadoHistorialDetalle struct {
	PedidoEstadoHistorial
	Estado  EstadoPedido   `json:"estado"`
	Usuario UsuarioResumen `json:"usuario"`
}
