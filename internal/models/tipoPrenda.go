package models

import "time"

type TipoPrenda struct {
	ID          int       `json:"id"`
	Nombre      string    `json:"nombre"`
	Descripcion *string   `json:"descripcion,omitempty"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TipoPrendaCreateRequest struct {
	Nombre      string  `json:"nombre"`
	Descripcion *string `json:"descripcion"`
}

type TipoPrendaUpdateRequest struct {
	Nombre      *string `json:"nombre"`
	Descripcion *string `json:"descripcion"`
	Activo      *bool   `json:"activo"`
}
