package models

import "time"

// PrendaCreateRequest contiene los datos de una prenda que se registra en
// un pedido existente. El pedido se toma de la URL del endpoint.
type PrendaCreateRequest struct {
	TipoPrendaID int     `json:"tipo_prenda_id"`
	Descripcion  *string `json:"descripcion,omitempty"`
	Cantidad     int     `json:"cantidad,omitempty"`
	Color        *string `json:"color,omitempty"`
}

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
