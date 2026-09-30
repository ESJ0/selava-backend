package models

// ReporteConsumoInsumos resume los movimientos de salida registrados durante
// un periodo y presenta el consumo acumulado de cada insumo.
type ReporteConsumoInsumos struct {
	FechaInicio         string                        `json:"fecha_inicio"`
	FechaFin            string                        `json:"fecha_fin"`
	CantidadMovimientos int                           `json:"cantidad_movimientos"`
	TotalConsumido      float64                       `json:"total_consumido"`
	ConsumoPorInsumo    []ReporteConsumoInsumoDetalle `json:"consumo_por_insumo"`
}

type ReporteConsumoInsumoDetalle struct {
	InsumoID            int     `json:"insumo_id"`
	Nombre              string  `json:"nombre"`
	UnidadMedida        string  `json:"unidad_medida"`
	CantidadMovimientos int     `json:"cantidad_movimientos"`
	CantidadConsumida   float64 `json:"cantidad_consumida"`
}
