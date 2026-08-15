package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type MetodoPagoController struct {
	service *service.MetodoPagoService
}

func NewMetodoPagoController(service *service.MetodoPagoService) *MetodoPagoController {
	return &MetodoPagoController{service: service}
}

func (mc *MetodoPagoController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.MetodoPagoCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	metodo, err := mc.service.CrearMetodoPago(ctx, &req)
	if err != nil {
		mc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, metodo)
}

func (mc *MetodoPagoController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	metodo, err := mc.service.ObtenerMetodoPago(ctx, id)
	if err != nil {
		mc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, metodo)
}

func (mc *MetodoPagoController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	metodos, err := mc.service.ListarMetodosPago(ctx)
	if err != nil {
		mc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, metodos)
}

func (mc *MetodoPagoController) Actualizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}

	var req models.MetodoPagoUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	metodo, err := mc.service.ActualizarMetodoPago(ctx, id, &req)
	if err != nil {
		mc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, metodo)
}

func (mc *MetodoPagoController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	if err := mc.service.EliminarMetodoPago(ctx, id); err != nil {
		mc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (mc *MetodoPagoController) handleServiceError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}
	switch {
	case errors.Is(err, repository.ErrMetodoPagoNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrNombreMetodoPagoEnUso):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
