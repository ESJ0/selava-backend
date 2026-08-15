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

type TipoPrendaController struct {
	service *service.TipoPrendaService
}

func NewTipoPrendaController(service *service.TipoPrendaService) *TipoPrendaController {
	return &TipoPrendaController{service: service}
}

func (tc *TipoPrendaController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.TipoPrendaCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	tipo, err := tc.service.CrearTipoPrenda(ctx, &req)
	if err != nil {
		tc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, tipo)
}

func (tc *TipoPrendaController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	tipo, err := tc.service.ObtenerTipoPrenda(ctx, id)
	if err != nil {
		tc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tipo)
}

func (tc *TipoPrendaController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	tipos, err := tc.service.ListarTiposPrenda(ctx)
	if err != nil {
		tc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tipos)
}

func (tc *TipoPrendaController) Actualizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}

	var req models.TipoPrendaUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	tipo, err := tc.service.ActualizarTipoPrenda(ctx, id, &req)
	if err != nil {
		tc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tipo)
}

func (tc *TipoPrendaController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	if err := tc.service.EliminarTipoPrenda(ctx, id); err != nil {
		tc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (tc *TipoPrendaController) handleServiceError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}
	switch {
	case errors.Is(err, repository.ErrTipoPrendaNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrNombreTipoPrendaEnUso):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
