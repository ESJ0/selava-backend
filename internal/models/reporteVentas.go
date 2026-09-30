package models

// ReporteVentas resume los pagos registrados dentro de un periodo. En este
// reporte una venta es un pedido distinto que recibio al menos un pago.
type ReporteVentas struct {
	FechaInicio         string                       `json:"fecha_inicio"`
	FechaFin            string                       `json:"fecha_fin"`
	CantidadVentas      int                          `json:"cantidad_ventas"`
	CantidadPagos       int                          `json:"cantidad_pagos"`
	TotalVentas         float64                      `json:"total_ventas"`
	VentasPorDia        []ReporteVentasPorDia        `json:"ventas_por_dia"`
	VentasPorMetodoPago []ReporteVentasPorMetodoPago `json:"ventas_por_metodo_pago"`
}

// ReporteVentasPorDia agrupa los pagos por fecha calendario.
type ReporteVentasPorDia struct {
	Fecha          string  `json:"fecha"`
	CantidadVentas int     `json:"cantidad_ventas"`
	CantidadPagos  int     `json:"cantidad_pagos"`
	TotalVentas    float64 `json:"total_ventas"`
}

// ReporteVentasPorMetodoPago agrupa los pagos por el metodo utilizado.
type ReporteVentasPorMetodoPago struct {
	MetodoPago     MetodoPago `json:"metodo_pago"`
	CantidadVentas int        `json:"cantidad_ventas"`
	CantidadPagos  int        `json:"cantidad_pagos"`
	TotalVentas    float64    `json:"total_ventas"`
}
