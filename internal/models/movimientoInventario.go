package models

import "time"

// MovimientoInventario representa un registro de la tabla
// movimientos_inventario. InsumoID y UsuarioID corresponden a las relaciones
// con las tablas insumos y usuarios, respectivamente.
type MovimientoInventario struct {
	ID              int       `json:"id"`
	InsumoID        int       `json:"insumo_id"`
	UsuarioID       int       `json:"usuario_id"`
	TipoMovimiento  string    `json:"tipo_movimiento"`
	Cantidad        float64   `json:"cantidad"`
	Motivo          *string   `json:"motivo,omitempty"`
	FechaMovimiento time.Time `json:"fecha_movimiento"`
	CreatedAt       time.Time `json:"created_at"`
}

// MovimientoInventarioCreateRequest contiene los datos enviados al registrar
// un movimiento. El usuario responsable se obtiene del token autenticado.
type MovimientoInventarioCreateRequest struct {
	InsumoID       int     `json:"insumo_id"`
	TipoMovimiento string  `json:"tipo_movimiento"`
	Cantidad       float64 `json:"cantidad"`
	Motivo         *string `json:"motivo"`
}
