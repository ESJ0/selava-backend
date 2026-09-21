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

type InsumoController struct {
	service *service.InsumoService
}

func NewInsumoController(service *service.InsumoService) *InsumoController {
	return &InsumoController{service: service}
}

func (ic *InsumoController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.InsumoCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	insumo, err := ic.service.CrearInsumo(ctx, &req)
	if err != nil {
		ic.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, insumo)
}

func (ic *InsumoController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	insumo, err := ic.service.ObtenerInsumo(ctx, id)
	if err != nil {
		ic.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, insumo)
}

func (ic *InsumoController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	insumos, err := ic.service.ListarInsumos(ctx)
	if err != nil {
		ic.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, insumos)
}

func (ic *InsumoController) ListarBajoStock(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	insumos, err := ic.service.ListarInsumosBajoStock(ctx)
	if err != nil {
		ic.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, insumos)
}

func (ic *InsumoController) Actualizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}

	var req models.InsumoUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	insumo, err := ic.service.ActualizarInsumo(ctx, id, &req)
	if err != nil {
		ic.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, insumo)
}

func (ic *InsumoController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "id invalido")
		return
	}
	if err := ic.service.EliminarInsumo(ctx, id); err != nil {
		ic.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (ic *InsumoController) handleServiceError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}
	switch {
	case errors.Is(err, repository.ErrInsumoNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrNombreInsumoEnUso):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
