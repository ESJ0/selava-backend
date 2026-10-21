package models

import "time"

type PromocionRequest struct {
	Nombre            string         `json:"nombre"`
	Descripcion       string         `json:"descripcion"`
	TipoPrendaID      int            `json:"tipo_prenda_id"`
	ModalidadCobro    ModalidadCobro `json:"modalidad_cobro"`
	MedidaRequerida   float64        `json:"medida_requerida"`
	PrecioPromocional float64        `json:"precio_promocional"`
	FechaInicio       time.Time      `json:"fecha_inicio"`
	FechaFin          time.Time      `json:"fecha_fin"`
	Activo            bool           `json:"activo"`
}

type Promocion struct {
	PromocionRequest
	ID               int       `json:"id"`
	TipoPrendaNombre string    `json:"tipo_prenda_nombre"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// El navegador solo elige identificadores. No recibe autoridad sobre importes.
type PromocionSeleccion struct {
	TipoPrendaID int `json:"tipo_prenda_id"`
	PromocionID  int `json:"promocion_id"`
}

type PromocionAplicada struct {
	PromocionID            int            `json:"promocion_id"`
	Nombre                 string         `json:"nombre"`
	Descripcion            string         `json:"descripcion"`
	TipoPrendaID           int            `json:"tipo_prenda_id"`
	TipoPrendaNombre       string         `json:"tipo_prenda_nombre"`
	ModalidadCobroAplicada ModalidadCobro `json:"modalidad_cobro_aplicada"`
	MedidaRequerida        float64        `json:"medida_requerida"`
	PrecioPromocional      float64        `json:"precio_promocional"`
	MedidaDisponible       float64        `json:"medida_disponible"`
	Aplicaciones           int64          `json:"aplicaciones"`
	MedidaCubierta         float64        `json:"medida_cubierta"`
	Sobrante               float64        `json:"sobrante"`
	PrecioNormalAplicado   float64        `json:"precio_normal_aplicado"`
	SubtotalPromocional    float64        `json:"subtotal_promocional"`
	SubtotalNormal         float64        `json:"subtotal_normal"`
	SubtotalBase           float64        `json:"subtotal_base"`
	SubtotalTotal          float64        `json:"subtotal_total"`
	PrendaIDs              []int          `json:"prenda_ids,omitempty"`
}

type PromocionCotizacionGrupo struct {
	TipoPrendaID     int                 `json:"tipo_prenda_id"`
	TipoPrendaNombre string              `json:"tipo_prenda_nombre"`
	ModalidadCobro   ModalidadCobro      `json:"modalidad_cobro"`
	Medida           float64             `json:"medida"`
	PrecioNormal     float64             `json:"precio_normal"`
	SubtotalBase     float64             `json:"subtotal_base"`
	SubtotalTotal    float64             `json:"subtotal_total"`
	Elegibles        []PromocionAplicada `json:"elegibles"`
	Aplicada         *PromocionAplicada  `json:"aplicada"`
}

type PromocionCotizacion struct {
	Grupos []PromocionCotizacionGrupo `json:"grupos"`
	Total  float64                    `json:"total"`
}
