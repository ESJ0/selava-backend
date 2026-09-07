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

type ServicioController struct {
	service *service.ServicioService
}

func NewServicioController(service *service.ServicioService) *ServicioController {
	return &ServicioController{service: service}
}

func (sc *ServicioController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.ServicioCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	servicio, err := sc.service.CrearServicio(ctx, &req)
	if err != nil {
		sc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, servicio)
}

func (sc *ServicioController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	servicio, err := sc.service.ObtenerServicio(ctx, id)
	if err != nil {
		sc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, servicio)
}

func (sc *ServicioController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	servicios, err := sc.service.ListarServicios(ctx)
	if err != nil {
		sc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, servicios)
}

func (sc *ServicioController) Actualizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	var req models.ServicioUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	servicio, err := sc.service.ActualizarServicio(ctx, id, &req)
	if err != nil {
		sc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, servicio)
}

func (sc *ServicioController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	if err := sc.service.EliminarServicio(ctx, id); err != nil {
		sc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (sc *ServicioController) handleServiceError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}
	switch {
	case errors.Is(err, repository.ErrServicioNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrNombreServicioEnUso):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
