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

type PrendaController struct {
	service *service.PrendaService
}

func NewPrendaController(service *service.PrendaService) *PrendaController {
	return &PrendaController{service: service}
}

func (pc *PrendaController) CrearVarias(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	pedidoID, err := strconv.Atoi(chi.URLParam(r, "pedidoID"))
	if err != nil || pedidoID <= 0 {
		respondError(w, http.StatusBadRequest, "id de pedido invalido")
		return
	}

	var reqs []models.PrendaCreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		respondError(w, http.StatusBadRequest, "el cuerpo debe ser un arreglo de prendas valido")
		return
	}

	prendas, err := pc.service.RegistrarPrendas(ctx, pedidoID, reqs)
	if err != nil {
		pc.handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, prendas)
}

func (pc *PrendaController) handleServiceError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validationErrors})
		return
	}

	switch {
	case errors.Is(err, repository.ErrPedidoNoEncontrado):
		respondError(w, http.StatusNotFound, "pedido no encontrado")
	case errors.Is(err, repository.ErrTipoPrendaNoEncontrado):
		respondError(w, http.StatusNotFound, "tipo de prenda no encontrado")
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
