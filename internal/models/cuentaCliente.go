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

// TieneEmailVerificado indica si el cliente completo la verificacion de su
// correo de acceso. El detalle del token de verificacion pertenece al flujo de
// autenticacion, no a esta entidad persistente.
func (c CuentaCliente) TieneEmailVerificado() bool {
	return c.EmailVerificadoAt != nil
}
