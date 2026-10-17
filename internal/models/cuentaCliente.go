package models

import "time"

// CuentaCliente contiene las credenciales y el estado de acceso al portal de
// un cliente. ClienteID es unico en base de datos: un cliente puede no tener
// cuenta, pero nunca puede tener mas de una.
type CuentaCliente struct {
	ID                int        `json:"id"`
	ClienteID         int        `json:"cliente_id"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"-"`
	Activo            bool       `json:"activo"`
	EmailVerificadoAt *time.Time `json:"email_verificado_at,omitempty"`
	UltimoAccesoAt    *time.Time `json:"ultimo_acceso_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ClienteRegistroRequest contiene el perfil y las credenciales que la landing
// publica necesita para registrar un nuevo cliente. Password solo existe en la
// entrada de la API; nunca se persiste ni se incluye en una respuesta.
type ClienteRegistroRequest struct {
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
	Telefono  string `json:"telefono"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
	Password  string `json:"password"`
}

type ClienteLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ClienteRegistrado agrupa las dos filas creadas dentro de la misma
// transaccion de registro.
type ClienteRegistrado struct {
	Cliente Cliente       `json:"cliente"`
	Cuenta  CuentaCliente `json:"cuenta"`
}

type CuentaClienteResponse struct {
	ID                int        `json:"id"`
	ClienteID         int        `json:"cliente_id"`
	Email             string     `json:"email"`
	Activo            bool       `json:"activo"`
	EmailVerificadoAt *time.Time `json:"email_verificado_at,omitempty"`
	UltimoAccesoAt    *time.Time `json:"ultimo_acceso_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type ClienteRegistroResponse struct {
	Cliente ClienteResponse       `json:"cliente"`
	Cuenta  CuentaClienteResponse `json:"cuenta"`
}

// TieneEmailVerificado indica si el cliente completo la verificacion de su
// correo de acceso. El detalle del token de verificacion pertenece al flujo de
// autenticacion, no a esta entidad persistente.
func (c CuentaCliente) TieneEmailVerificado() bool {
	return c.EmailVerificadoAt != nil
}

func (c CuentaCliente) ToResponse() CuentaClienteResponse {
	return CuentaClienteResponse{
		ID:                c.ID,
		ClienteID:         c.ClienteID,
		Email:             c.Email,
		Activo:            c.Activo,
		EmailVerificadoAt: c.EmailVerificadoAt,
		UltimoAccesoAt:    c.UltimoAccesoAt,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

func (r ClienteRegistrado) ToResponse() ClienteRegistroResponse {
	return ClienteRegistroResponse{
		Cliente: r.Cliente.ToResponse(),
		Cuenta:  r.Cuenta.ToResponse(),
	}
}
