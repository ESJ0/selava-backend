package models

import "time"

// Insumo representa un registro de la tabla insumos.
type Insumo struct {
	ID           int       `json:"id"`
	Nombre       string    `json:"nombre"`
	Descripcion  *string   `json:"descripcion,omitempty"`
	UnidadMedida string    `json:"unidad_medida"`
	StockActual  float64   `json:"stock_actual"`
	StockMinimo  float64   `json:"stock_minimo"`
	Activo       bool      `json:"activo"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InsumoCreateRequest contiene los datos necesarios para registrar un insumo.
// Los stocks son opcionales y la base de datos los inicializa en cero cuando
// no se envian.
type InsumoCreateRequest struct {
	Nombre       string  `json:"nombre"`
	Descripcion  *string `json:"descripcion"`
	UnidadMedida string  `json:"unidad_medida"`
	StockActual  float64 `json:"stock_actual"`
	StockMinimo  float64 `json:"stock_minimo"`
}

// InsumoUpdateRequest permite modificar parcialmente un insumo.
// Los punteros permiten distinguir un campo omitido de su valor cero.
type InsumoUpdateRequest struct {
	Nombre       *string  `json:"nombre"`
	Descripcion  *string  `json:"descripcion"`
	UnidadMedida *string  `json:"unidad_medida"`
	StockActual  *float64 `json:"stock_actual"`
	StockMinimo  *float64 `json:"stock_minimo"`
	Activo       *bool    `json:"activo"`
}
