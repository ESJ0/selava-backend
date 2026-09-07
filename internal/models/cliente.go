package models

import "time"

// Cliente representa un registro de la tabla clientes.
type Cliente struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Apellido  string    `json:"apellido"`
	Telefono  string    `json:"telefono"`
	Email     string    `json:"email,omitempty"`
	Direccion string    `json:"direccion,omitempty"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Payload que se recibe al crear un cliente
type ClienteCreateRequest struct {
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
	Telefono  string `json:"telefono"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

// Payload que se recibe al actualizar un cliente (PATCH/PUT)
type ClienteUpdateRequest struct {
	Nombre    *string `json:"nombre"`
	Apellido  *string `json:"apellido"`
	Telefono  *string `json:"telefono"`
	Email     *string `json:"email"`
	Direccion *string `json:"direccion"`
	Activo    *bool   `json:"activo"`
}

// Es lo que se expone hacia afuera en la API
type ClienteResponse struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Apellido  string    `json:"apellido"`
	Telefono  string    `json:"telefono"`
	Email     string    `json:"email,omitempty"`
	Direccion string    `json:"direccion,omitempty"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Convierte un Cliente en su versión de respuesta para la API
func (c *Cliente) ToResponse() ClienteResponse {
	return ClienteResponse{
		ID:        c.ID,
		Nombre:    c.Nombre,
		Apellido:  c.Apellido,
		Telefono:  c.Telefono,
		Email:     c.Email,
		Direccion: c.Direccion,
		Activo:    c.Activo,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
