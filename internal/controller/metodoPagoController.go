package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
	"net/http"
)

type MetodoPagoController struct{ service *service.MetodoPagoService }

func NewMetodoPagoController(s *service.MetodoPagoService) *MetodoPagoController {
	return &MetodoPagoController{service: s}
}
func (mc *MetodoPagoController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req models.MetodoPagoCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		respondError(w, 400, "cuerpo de la peticion invalido")
		return
	}
	m, err := mc.service.CrearMetodoPago(ctx, &req)
	if err != nil {
		mc.handle(w, err)
		return
	}
	respondJSON(w, 201, m)
}
func (mc *MetodoPagoController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, 400, "id invalido")
		return
	}
	m, err := mc.service.ObtenerMetodoPago(ctx, id)
	if err != nil {
		mc.handle(w, err)
		return
	}
	respondJSON(w, 200, m)
}
func (mc *MetodoPagoController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	m, err := mc.service.ListarMetodosPago(ctx)
	if err != nil {
		mc.handle(w, err)
		return
	}
	respondJSON(w, 200, m)
}
func (mc *MetodoPagoController) Actualizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, 400, "id invalido")
		return
	}
	var req models.MetodoPagoUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		respondError(w, 400, "cuerpo de la peticion invalido")
		return
	}
	m, err := mc.service.ActualizarMetodoPago(ctx, id, &req)
	if err != nil {
		mc.handle(w, err)
		return
	}
	respondJSON(w, 200, m)
}
func (mc *MetodoPagoController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, 400, "id invalido")
		return
	}
	if err := mc.service.EliminarMetodoPago(ctx, id); err != nil {
		mc.handle(w, err)
		return
	}
	respondJSON(w, 204, nil)
}
func (mc *MetodoPagoController) handle(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		respondJSON(w, 422, map[string]any{"errores": ve})
		return
	}
	switch {
	case errors.Is(err, repository.ErrMetodoPagoNoEncontrado):
		respondError(w, 404, err.Error())
	case errors.Is(err, repository.ErrNombreMetodoPagoEnUso):
		respondError(w, 409, err.Error())
	default:
		respondError(w, 500, "error interno del servidor")
	}
}
