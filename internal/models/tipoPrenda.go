package models

import "time"

type ModalidadCobro string

const (
	ModalidadCobroUnidad ModalidadCobro = "UNIDAD"
	ModalidadCobroLibra  ModalidadCobro = "LIBRA"
	ModalidadCobroMetro  ModalidadCobro = "METRO"
)

type TipoPrenda struct {
	Precio         *float64       `json:"precio"`
	ModalidadCobro ModalidadCobro `json:"modalidad_cobro"`
	ID             int            `json:"id"`
	Nombre         string         `json:"nombre"`
	Descripcion    *string        `json:"descripcion,omitempty"`
	Activo         bool           `json:"activo"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type TipoPrendaCreateRequest struct {
	Precio         *float64       `json:"precio"`
	ModalidadCobro ModalidadCobro `json:"modalidad_cobro"`
	Nombre         string         `json:"nombre"`
	Descripcion    *string        `json:"descripcion"`
}

type TipoPrendaUpdateRequest struct {
	Precio         *float64        `json:"precio"`
	ModalidadCobro *ModalidadCobro `json:"modalidad_cobro"`
	Nombre         *string         `json:"nombre"`
	Descripcion    *string         `json:"descripcion"`
	Activo         *bool           `json:"activo"`
}

// TipoPrendaPublico no expone estado ni datos administrativos.
type TipoPrendaPublico struct {
	ID             int            `json:"id"`
	Nombre         string         `json:"nombre"`
	Descripcion    *string        `json:"descripcion"`
	Precio         float64        `json:"precio"`
	ModalidadCobro ModalidadCobro `json:"modalidad_cobro"`
}
