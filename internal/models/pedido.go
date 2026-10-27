package models

import "time"

// Pedido representa un registro de la tabla pedidos.
type Pedido struct {
	ID                   int        `json:"id"`
	ClienteID            int        `json:"cliente_id"`
	UsuarioID            int        `json:"usuario_id"`
	EstadoActualID       int        `json:"estado_actual_id"`
	FechaRecibido        time.Time  `json:"fecha_recibido"`
	FechaEntregaEstimada *time.Time `json:"fecha_entrega_estimada,omitempty"`
	FechaEntregaReal     *time.Time `json:"fecha_entrega_real,omitempty"`
	Total                float64    `json:"total"`
	Observaciones        *string    `json:"observaciones,omitempty"`
	Activo               bool       `json:"activo"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PedidoCreateRequest contiene los datos que puede enviar el cliente al
// registrar un pedido. El usuario y el estado inicial los determina el
// backend a partir del token y del catalogo de estados, respectivamente.
type PedidoCreateRequest struct {
	Promociones          []PromocionSeleccion  `json:"promociones,omitempty"`
	ClienteID            int                   `json:"cliente_id"`
	FechaEntregaEstimada *time.Time            `json:"fecha_entrega_estimada,omitempty"`
	Observaciones        *string               `json:"observaciones,omitempty"`
	Prendas              []PrendaCreateRequest `json:"prendas"`
}

// PedidoConPrendas es la respuesta que se devuelve al crear un pedido:
// el pedido junto con las prendas que quedaron registradas.
type PedidoConPrendas struct {
	Pedido
	PromocionesAplicadas []PromocionAplicada `json:"promociones_aplicadas"`
	Prendas              []Prenda            `json:"prendas"`
}

// EstadoPedido representa el estado del catalogo asociado a un pedido.
type EstadoPedido struct {
	ID            int       `json:"id"`
	Nombre        string    `json:"nombre"`
	Codigo        string    `json:"codigo"`
	Orden         int       `json:"orden"`
	EsObligatorio bool      `json:"es_obligatorio"`
	EsTerminal    bool      `json:"es_terminal"`
	EsCancelacion bool      `json:"es_cancelacion"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ServicioEtapa resume el servicio contratado que hace aplicable una etapa.
type ServicioEtapa struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

// EstadoFlujoPedido agrega el avance y los servicios que originan cada etapa.
type EstadoFlujoPedido struct {
	EstadoPedido
	Servicios  []ServicioEtapa `json:"servicios"`
	Completado bool            `json:"completado"`
	Actual     bool            `json:"actual"`
}

// FlujoOperativoPedido es el plan inmutable en significado, pero calculado a
// partir de los servicios que realmente fueron contratados por el pedido.
type FlujoOperativoPedido struct {
	Estados            []EstadoFlujoPedido `json:"estados"`
	SiguienteEstado    *EstadoPedido       `json:"siguiente_estado,omitempty"`
	ProgresoPorcentaje int                 `json:"progreso_porcentaje"`
	Cancelado          bool                `json:"cancelado"`
}

// PedidoDetalle contiene toda la informacion relacionada con un pedido.
type PedidoDetalle struct {
	Pedido
	PromocionesAplicadas []PromocionAplicada  `json:"promociones_aplicadas"`
	Cliente              Cliente              `json:"cliente"`
	Prendas              []Prenda             `json:"prendas"`
	EstadoActual         EstadoPedido         `json:"estado_actual"`
	Pagos                []PagoDetalle        `json:"pagos"`
	FlujoOperativo       FlujoOperativoPedido `json:"flujo_operativo"`
}
