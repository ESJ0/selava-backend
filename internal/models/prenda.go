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

// PrendaCreateRequest contiene los datos de una prenda enviados como parte
// de la creacion de un pedido. El pedido_id lo asigna el backend una vez
// creado el pedido padre.
type PrendaCreateRequest struct {
	TipoPrendaID int     `json:"tipo_prenda_id"`
	Descripcion  *string `json:"descripcion,omitempty"`
	Cantidad     int     `json:"cantidad"`
	Color        *string `json:"color,omitempty"`
}
