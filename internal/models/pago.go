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

// PagoCreateRequest contiene los datos aceptados al registrar un abono.
// El pedido se obtiene de la URL y el usuario del token autenticado.
type PagoCreateRequest struct {
	MetodoPagoID int     `json:"metodo_pago_id"`
	Monto        float64 `json:"monto"`
	Referencia   *string `json:"referencia,omitempty"`
}

// SaldoPedido resume el total, los abonos acumulados y lo que falta pagar.
type SaldoPedido struct {
	PedidoID       int     `json:"pedido_id"`
	Total          float64 `json:"total"`
	TotalPagado    float64 `json:"total_pagado"`
	SaldoPendiente float64 `json:"saldo_pendiente"`
}

// PagoUsuario identifica de forma segura a quien registro el pago.
type PagoUsuario struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
}

// PagoDetalle incluye el metodo de pago utilizado.
type PagoDetalle struct {
	Pago
	MetodoPago MetodoPago   `json:"metodo_pago"`
	Usuario    *PagoUsuario `json:"usuario,omitempty"`
}
