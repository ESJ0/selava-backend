package models

import "time"

// PrendaServicio representa el servicio aplicado a una prenda y conserva
// el precio vigente en el momento en que se creo la relacion.
type PrendaServicio struct {
	ID             int       `json:"id"`
	PrendaID       int       `json:"prenda_id"`
	ServicioID     int       `json:"servicio_id"`
	Servicio       *Servicio `json:"servicio,omitempty"`
	PrecioAplicado float64   `json:"precio_aplicado"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
