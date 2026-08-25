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

type PedidoController struct {
	service *service.PedidoService
}

func NewPedidoController(service *service.PedidoService) *PedidoController {
	return &PedidoController{service: service}
}

func (pc *PedidoController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims.UsuarioID <= 0 {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return
	}

	var req models.PedidoCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}

	pedido, err := pc.service.CrearPedido(ctx, &req, claims.UsuarioID)
	if err != nil {
		pc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, pedido)
}

func (pc *PedidoController) handleServiceError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}

	switch {
	case errors.Is(err, repository.ErrClienteNoEncontrado):
		respondError(w, http.StatusNotFound, "cliente no encontrado")
	case errors.Is(err, repository.ErrUsuarioNoEncontrado):
		respondError(w, http.StatusUnauthorized, "usuario autenticado no encontrado")
	case errors.Is(err, repository.ErrTipoPrendaNoEncontrado):
		respondError(w, http.StatusNotFound, "tipo de prenda no encontrado")
	case errors.Is(err, repository.ErrServicioNoEncontrado):
		respondError(w, http.StatusNotFound, "servicio no encontrado o inactivo")
	case errors.Is(err, repository.ErrEstadoPedidoNoEncontrado):
		respondError(w, http.StatusInternalServerError, "estado inicial de pedido no configurado")
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
