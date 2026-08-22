package models

import "time"

// Prenda representa un registro de la tabla prendas.
type Prenda struct {
	ID           int       `json:"id"`
	PedidoID     int       `json:"pedido_id"`
	TipoPrendaID int       `json:"tipo_prenda_id"`
	Descripcion  *string   `json:"descripcion,omitempty"`
	Cantidad     int       `json:"cantidad"`
	Color        *string   `json:"color,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
