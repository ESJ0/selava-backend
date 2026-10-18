package models

import "time"

type PedidoListRequest struct {
	Buscar, EstadoID, FechaDesde, FechaHasta, Pagina, Limite, Orden string
}

type PedidoListFilter struct {
	ClienteID           int
	Buscar              string
	EstadoID            int
	FechaDesde          *time.Time
	FechaHastaExclusiva *time.Time
	Pagina              int
	Limite              int
	Orden               string
}

// PedidoPropioResumen contiene solo la informacion que necesita el portal
// cliente. Evita exponer datos internos del empleado que registro el pedido.
type PedidoPropioResumen struct {
	ID                   int                 `json:"id"`
	FechaRecibido        time.Time           `json:"fecha_recibido"`
	FechaEntregaEstimada *time.Time          `json:"fecha_entrega_estimada,omitempty"`
	Total                float64             `json:"total"`
	EstadoActual         PedidoEstadoResumen `json:"estado_actual"`
}

type PedidosPropiosResponse struct {
	Pedidos []PedidoPropioResumen `json:"pedidos"`
	Total   int                   `json:"total"`
	Pagina  int                   `json:"pagina"`
	Limite  int                   `json:"limite"`
}

type PedidoClienteResumen struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Telefono string `json:"telefono"`
}

type PedidoEstadoResumen struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Orden  int    `json:"orden"`
}

type PedidoResumen struct {
	ID                   int                  `json:"id"`
	FechaRecibido        time.Time            `json:"fecha_recibido"`
	FechaEntregaEstimada *time.Time           `json:"fecha_entrega_estimada,omitempty"`
	Total                float64              `json:"total"`
	Activo               bool                 `json:"activo"`
	Cliente              PedidoClienteResumen `json:"cliente"`
	EstadoActual         PedidoEstadoResumen  `json:"estado_actual"`
}

type PedidoListResponse struct {
	Pedidos []PedidoResumen `json:"pedidos"`
	Total   int             `json:"total"`
	Pagina  int             `json:"pagina"`
	Limite  int             `json:"limite"`
}
