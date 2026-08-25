package models

type MetodoPago struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
}

type MetodoPagoCreateRequest struct {
	Nombre string `json:"nombre"`
}
type MetodoPagoUpdateRequest struct {
	Nombre *string `json:"nombre"`
	Activo *bool   `json:"activo"`
}
