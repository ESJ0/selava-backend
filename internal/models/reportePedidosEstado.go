package models

// ReportePedidosEstado resume los pedidos activos segun su estado actual.
type ReportePedidosEstado struct {
	TotalPedidos     int                           `json:"total_pedidos"`
	PedidosPorEstado []ReportePedidosEstadoDetalle `json:"pedidos_por_estado"`
}

// ReportePedidosEstadoDetalle contiene el total asociado a un estado del
// catalogo. Los estados sin pedidos se incluyen con cantidad cero.
type ReportePedidosEstadoDetalle struct {
	Estado          EstadoPedido `json:"estado"`
	CantidadPedidos int          `json:"cantidad_pedidos"`
}
