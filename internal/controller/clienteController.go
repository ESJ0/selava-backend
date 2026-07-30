package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

// Timeout corto: este endpoint se usa en el punto de venta mientras se
// atiende al cliente, así que una consulta colgada no puede bloquear la caja.
const requestTimeout = 3 * time.Second

const maxBodyBytes = 1 << 20 // 1MB, de sobra para un payload de cliente

type ClienteController struct {
	service *service.ClienteService
}

func NewClienteController(service *service.ClienteService) *ClienteController {
	return &ClienteController{service: service}
}

func (cc *ClienteController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.ClienteCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	cliente, err := cc.service.CrearCliente(ctx, &req)
	if err != nil {
		cc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, cliente.ToResponse())
}

func (cc *ClienteController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id inválido")
		return
	}

	cliente, err := cc.service.ObtenerCliente(ctx, id)
	if err != nil {
		cc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, cliente.ToResponse())
}

func (cc *ClienteController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	clientes, err := cc.service.ListarClientes(ctx)
	if err != nil {
		cc.handleServiceError(w, err)
		return
	}

	respuesta := make([]models.ClienteResponse, 0, len(clientes))
	for i := range clientes {
		respuesta = append(respuesta, clientes[i].ToResponse())
	}
	respondJSON(w, http.StatusOK, respuesta)
}

func (cc *ClienteController) Actualizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req models.ClienteUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	cliente, err := cc.service.ActualizarCliente(ctx, id, &req)
	if err != nil {
		cc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, cliente.ToResponse())
}

func (cc *ClienteController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := cc.service.EliminarCliente(ctx, id); err != nil {
		cc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (cc *ClienteController) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrClienteNoEncontrado):
		respondError(w, http.StatusNotFound, "cliente no encontrado")
	case errors.Is(err, service.ErrEmailEnUso):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}
