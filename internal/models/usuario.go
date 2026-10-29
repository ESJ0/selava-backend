package models

import "time"

// Usuario representa un registro de la tabla usuarios.
type Usuario struct {
	ID                int        `json:"id"`
	RolID             int        `json:"rol_id"`
	Rol               string     `json:"rol"`
	ClienteID         *int       `json:"cliente_id,omitempty"`
	Nombre            string     `json:"nombre"`
	Apellido          string     `json:"apellido"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"-"`
	Telefono          string     `json:"telefono"`
	Activo            bool       `json:"activo"`
	UltimoAccesoAt    *time.Time `json:"ultimo_acceso_at,omitempty"`
	EmailVerificadoAt *time.Time `json:"email_verificado_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Payload que se recibe al crear un usuario (contraseñasin hashear)
type UsuarioCreateRequest struct {
	RolID     int    `json:"rol_id"`
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Telefono  string `json:"telefono"`
	Direccion string `json:"direccion"`
}

type UsuarioUpdateRequest struct {
	Nombre    *string `json:"nombre"`
	Apellido  *string `json:"apellido"`
	Email     *string `json:"email"`
	Telefono  *string `json:"telefono"`
	Direccion *string `json:"direccion"`
	Activo    *bool   `json:"activo"`
}

type LoginUnificadoResponse struct {
	Token   string          `json:"token"`
	Usuario UsuarioResponse `json:"usuario"`
	Destino string          `json:"destino"`
}

// Es lo que se expone hacia afuera en la API
type UsuarioResponse struct {
	ID        int       `json:"id"`
	RolID     int       `json:"rol_id"`
	Rol       string    `json:"rol"`
	ClienteID *int      `json:"cliente_id,omitempty"`
	Nombre    string    `json:"nombre"`
	Apellido  string    `json:"apellido"`
	Email     string    `json:"email"`
	Telefono  string    `json:"telefono"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Convierte un Usuario en su versión segura para exponer en la API
func (u *Usuario) ToResponse() UsuarioResponse {
	return UsuarioResponse{
		ID:        u.ID,
		RolID:     u.RolID,
		Rol:       u.Rol,
		ClienteID: u.ClienteID,
		Nombre:    u.Nombre,
		Apellido:  u.Apellido,
		Email:     u.Email,
		Telefono:  u.Telefono,
		Activo:    u.Activo,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
