package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type MovimientoInventarioController struct {
	service *service.MovimientoInventarioService
}

func NewMovimientoInventarioController(service *service.MovimientoInventarioService) *MovimientoInventarioController {
	return &MovimientoInventarioController{service: service}
}

func (mc *MovimientoInventarioController) Registrar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims.UsuarioID <= 0 {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return
	}

	var req models.MovimientoInventarioCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}

	movimiento, err := mc.service.RegistrarMovimiento(ctx, &req, claims.UsuarioID)
	if err != nil {
		mc.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, movimiento)
}

func (mc *MovimientoInventarioController) handleError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}
	switch {
	case errors.Is(err, repository.ErrInsumoNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrUsuarioNoEncontrado):
		respondError(w, http.StatusUnauthorized, "usuario autenticado no encontrado")
	case errors.Is(err, repository.ErrStockInsuficiente):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
