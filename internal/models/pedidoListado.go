package models

import "time"

type PedidoListRequest struct {
	Buscar, EstadoID, FechaDesde, FechaHasta, Pagina, Limite, Orden string
	// Filtros del portal de clientes.
	Seccion, TipoPrendaID, Saldo string
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
	// Seccion: "" (todos), "activos" (en proceso) o "historial" (finalizados o cancelados).
	Seccion      string
	TipoPrendaID int
	// Saldo: "" (todos), "pendiente" (con saldo, sin cancelados) o "pagado" (sin saldo).
	Saldo string
	// SoloVigentes excluye pedidos eliminados lógicamente (portal de clientes).
	SoloVigentes bool
}

// PedidoPropioResumen contiene solo la informacion que necesita el portal
// cliente. Evita exponer datos internos del empleado que registro el pedido.
type PedidoPropioResumen struct {
	ID                   int                 `json:"id"`
	FechaRecibido        time.Time           `json:"fecha_recibido"`
	FechaEntregaEstimada *time.Time          `json:"fecha_entrega_estimada,omitempty"`
	Total                float64             `json:"total"`
	TotalPagado          float64             `json:"total_pagado"`
	SaldoPendiente       float64             `json:"saldo_pendiente"`
	EnProceso            bool                `json:"en_proceso"`
	Cancelado            bool                `json:"cancelado"`
	TiposPrenda          []string            `json:"tipos_prenda"`
	CantidadPrendas      int                 `json:"cantidad_prendas"`
	EstadoActual         PedidoEstadoResumen `json:"estado_actual"`
}

// PedidosClienteTotales resume todos los pedidos vigentes del cliente (sin filtros).
type PedidosClienteTotales struct {
	Activos         int     `json:"activos"`
	Historial       int     `json:"historial"`
	PedidosConSaldo int     `json:"pedidos_con_saldo"`
	SaldoPendiente  float64 `json:"saldo_pendiente"`
	TotalPagado     float64 `json:"total_pagado"`
	TotalFacturado  float64 `json:"total_facturado"`
}

type PedidosPropiosResponse struct {
	Pedidos []PedidoPropioResumen  `json:"pedidos"`
	Total   int                    `json:"total"`
	Pagina  int                    `json:"pagina"`
	Limite  int                    `json:"limite"`
	Resumen *PedidosClienteTotales `json:"resumen,omitempty"`
}

// TipoPrendaOpcion es un tipo de prenda presente en los pedidos del cliente (para filtros).
type TipoPrendaOpcion struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
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
	TotalPagado          float64              `json:"total_pagado"`
	SaldoPendiente       float64              `json:"saldo_pendiente"`
	EnProceso            bool                 `json:"en_proceso"`
	Cancelado            bool                 `json:"cancelado"`
	TiposPrenda          []string             `json:"tipos_prenda"`
	CantidadPrendas      int                  `json:"cantidad_prendas"`
	Cliente              PedidoClienteResumen `json:"cliente"`
	EstadoActual         PedidoEstadoResumen  `json:"estado_actual"`
}

type PedidoListResponse struct {
	Pedidos []PedidoResumen        `json:"pedidos"`
	Total   int                    `json:"total"`
	Pagina  int                    `json:"pagina"`
	Limite  int                    `json:"limite"`
	Resumen *PedidosClienteTotales `json:"resumen,omitempty"`
}
