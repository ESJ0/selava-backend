package models

import "time"

// Pago representa un pago registrado para un pedido.
type Pago struct {
	ID           int       `json:"id"`
	PedidoID     int       `json:"pedido_id"`
	MetodoPagoID int       `json:"metodo_pago_id"`
	UsuarioID    int       `json:"usuario_id"`
	Monto        float64   `json:"monto"`
	Referencia   *string   `json:"referencia,omitempty"`
	FechaPago    time.Time `json:"fecha_pago"`
	CreatedAt    time.Time `json:"created_at"`
}

// PagoDetalle incluye el metodo de pago utilizado.
type PagoDetalle struct {
	Pago
	MetodoPago MetodoPago `json:"metodo_pago"`
}
