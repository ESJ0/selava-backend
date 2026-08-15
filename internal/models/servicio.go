package models

import "time"

// Servicio representa un servicio ofrecido por la lavanderia.
type Servicio struct {
	ID                  int       `json:"id"`
	Nombre              string    `json:"nombre"`
	Descripcion         *string   `json:"descripcion,omitempty"`
	PrecioBase          float64   `json:"precio_base"`
	TiempoEstimadoHoras *int      `json:"tiempo_estimado_horas,omitempty"`
	Activo              bool      `json:"activo"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ServicioCreateRequest struct {
	Nombre              string  `json:"nombre"`
	Descripcion         *string `json:"descripcion"`
	PrecioBase          float64 `json:"precio_base"`
	TiempoEstimadoHoras *int    `json:"tiempo_estimado_horas"`
}

// Los punteros permiten distinguir entre un campo omitido y su valor cero.
type ServicioUpdateRequest struct {
	Nombre              *string  `json:"nombre"`
	Descripcion         *string  `json:"descripcion"`
	PrecioBase          *float64 `json:"precio_base"`
	TiempoEstimadoHoras *int     `json:"tiempo_estimado_horas"`
	Activo              *bool    `json:"activo"`
}
