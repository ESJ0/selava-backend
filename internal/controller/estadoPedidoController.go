package controller

import (
	"context"
	"net/http"

	"github.com/ESJ0/selava-backend/internal/service"
)

type EstadoPedidoController struct {
	service *service.EstadoPedidoService
}

func NewEstadoPedidoController(service *service.EstadoPedidoService) *EstadoPedidoController {
	return &EstadoPedidoController{service: service}
}

func (ec *EstadoPedidoController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	estados, err := ec.service.ListarEstados(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	respondJSON(w, http.StatusOK, estados)
}
