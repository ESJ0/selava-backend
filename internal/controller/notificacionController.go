package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type NotificacionController struct {
	service *service.NotificacionService
}

func NewNotificacionController(service *service.NotificacionService) *NotificacionController {
	return &NotificacionController{service: service}
}

func (c *NotificacionController) ListarPropias(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	clienteID, ok := clienteIDAutenticado(w, r)
	if !ok {
		return
	}
	result, err := c.service.ListarPropias(ctx, clienteID)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (c *NotificacionController) MarcarLeida(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	clienteID, ok := clienteIDAutenticado(w, r)
	if !ok {
		return
	}
	notificacionID, err := strconv.Atoi(chi.URLParam(r, "notificacionID"))
	if err != nil || notificacionID <= 0 {
		respondError(w, http.StatusBadRequest, "id de notificacion invalido")
		return
	}
	result, err := c.service.MarcarLeida(ctx, notificacionID, clienteID)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (c *NotificacionController) MarcarTodasLeidas(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	clienteID, ok := clienteIDAutenticado(w, r)
	if !ok {
		return
	}
	actualizadas, err := c.service.MarcarTodasLeidas(ctx, clienteID)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, models.NotificacionesActualizadasResponse{Actualizadas: actualizadas})
}

func (c *NotificacionController) Eliminar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	clienteID, ok := clienteIDAutenticado(w, r)
	if !ok {
		return
	}
	notificacionID, err := strconv.Atoi(chi.URLParam(r, "notificacionID"))
	if err != nil || notificacionID <= 0 {
		respondError(w, http.StatusBadRequest, "id de notificacion invalido")
		return
	}
	if err := c.service.Eliminar(ctx, notificacionID, clienteID); err != nil {
		c.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func clienteIDAutenticado(w http.ResponseWriter, r *http.Request) (int, bool) {
	claims, ok := middleware.ClienteClaimsFromContext(r.Context())
	if !ok || claims.ClienteID <= 0 {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return 0, false
	}
	return claims.ClienteID, true
}

func (c *NotificacionController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotificacionNoEncontrada):
		respondError(w, http.StatusNotFound, "notificacion no encontrada")
	case errors.Is(err, repository.ErrClienteNoEncontrado):
		respondError(w, http.StatusUnauthorized, "cliente autenticado invalido")
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
