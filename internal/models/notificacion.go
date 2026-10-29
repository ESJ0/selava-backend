package models

import "time"

// Notificacion es un aviso privado del portal de clientes. PedidoID es
// opcional para permitir avisos generales y conservar el historial si un
// pedido se elimina posteriormente.
type Notificacion struct {
	ID        int       `json:"id"`
	ClienteID int       `json:"cliente_id"`
	PedidoID  *int      `json:"pedido_id,omitempty"`
	Mensaje   string    `json:"mensaje"`
	Leida     bool      `json:"leida"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificacionesResponse struct {
	Notificaciones []Notificacion `json:"notificaciones"`
	NoLeidas       int            `json:"no_leidas"`
}

type NotificacionesActualizadasResponse struct {
	Actualizadas int64 `json:"actualizadas"`
}
