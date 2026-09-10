package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

type PagoController struct{ service *service.PagoService }

func NewPagoController(service *service.PagoService) *PagoController {
	return &PagoController{service: service}
}

func (pc *PagoController) ObtenerSaldo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	pedidoID, err := strconv.Atoi(chi.URLParam(r, "pedidoID"))
	if err != nil || pedidoID <= 0 {
		respondError(w, http.StatusBadRequest, "id de pedido invalido")
		return
	}
	saldo, err := pc.service.ObtenerSaldo(ctx, pedidoID)
	if err != nil {
		pc.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, saldo)
}

func (pc *PagoController) ObtenerHistorial(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	pedidoID, err := strconv.Atoi(chi.URLParam(r, "pedidoID"))
	if err != nil || pedidoID <= 0 {
		respondError(w, http.StatusBadRequest, "id de pedido invalido")
		return
	}
	pagos, err := pc.service.ObtenerHistorial(ctx, pedidoID)
	if err != nil {
		pc.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, pagos)
}

func (pc *PagoController) Registrar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	pedidoID, err := strconv.Atoi(chi.URLParam(r, "pedidoID"))
	if err != nil || pedidoID <= 0 {
		respondError(w, http.StatusBadRequest, "id de pedido invalido")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims.UsuarioID <= 0 {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return
	}
	var req models.PagoCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	pago, err := pc.service.RegistrarPago(ctx, pedidoID, &req, claims.UsuarioID)
	if err != nil {
		pc.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, pago)
}

func (pc *PagoController) handleError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}
	switch {
	case errors.Is(err, repository.ErrPedidoNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrMetodoPagoNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrUsuarioNoEncontrado):
		respondError(w, http.StatusUnauthorized, "usuario autenticado no encontrado")
	case errors.Is(err, repository.ErrPagoExcedeSaldo), errors.Is(err, repository.ErrPedidoSinSaldo), errors.Is(err, repository.ErrPagoPedidoCancelado):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
