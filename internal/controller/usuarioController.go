package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

type UsuarioController struct{ service *service.UsuarioService }

func NewUsuarioController(service *service.UsuarioService) *UsuarioController {
	return &UsuarioController{service: service}
}

func (c *UsuarioController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	usuarios, err := c.service.Listar(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	respondJSON(w, http.StatusOK, usuarios)
}

func (c *UsuarioController) Crear(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req models.UsuarioCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	usuario, err := c.service.Crear(ctx, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, usuario)
}

func (c *UsuarioController) Actualizar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "usuarioID"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "usuario invalido")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req models.UsuarioUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	usuario, err := c.service.Actualizar(ctx, id, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, usuario)
}

func (c *UsuarioController) handleError(w http.ResponseWriter, err error) {
	var validation validator.ValidationErrors
	switch {
	case errors.As(err, &validation):
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validation})
	case errors.Is(err, repository.ErrEmailUsuarioEnUso):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, repository.ErrUsuarioNoEncontrado):
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
