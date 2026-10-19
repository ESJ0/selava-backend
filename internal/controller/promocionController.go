package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
	"io"
	"net/http"
)

type PromocionController struct{ service *service.PromocionService }

func NewPromocionController(s *service.PromocionService) *PromocionController {
	return &PromocionController{service: s}
}
func (c *PromocionController) handleError(w http.ResponseWriter, err error) {
	var fields validator.ValidationErrors
	if errors.As(err, &fields) {
		respondJSON(w, 422, map[string]any{"errores": fields})
		return
	}
	switch {
	case errors.Is(err, repository.ErrPromocionNoEncontrada), errors.Is(err, repository.ErrTipoPrendaNoEncontrado):
		respondError(w, 404, err.Error())
	case errors.Is(err, repository.ErrPromocionInvalida), errors.Is(err, repository.ErrTipoPrendaSinPrecio), errors.Is(err, repository.ErrMedidaInvalidaModalidad), errors.Is(err, repository.ErrTotalPedidoFueraDeRango):
		respondJSON(w, 422, map[string]any{"errores": validator.ValidationErrors{{Field: "promociones", Message: err.Error()}}})
	default:
		respondError(w, 500, "error interno del servidor")
	}
}
func decodePromocionBody(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		respondError(w, 400, "cuerpo de la peticion invalido")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		respondError(w, 400, "cuerpo de la peticion invalido")
		return false
	}
	return true
}
func (c *PromocionController) Listar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	result, err := c.service.List(ctx)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, 200, result)
}
func (c *PromocionController) Obtener(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, 400, "id invalido")
		return
	}
	result, err := c.service.Get(ctx, id)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, 200, result)
}
func (c *PromocionController) Crear(w http.ResponseWriter, r *http.Request) { c.save(w, r, 0) }
func (c *PromocionController) Actualizar(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, 400, "id invalido")
		return
	}
	c.save(w, r, id)
}
func (c *PromocionController) save(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req models.PromocionRequest
	if !decodePromocionBody(w, r, &req) {
		return
	}
	result, err := c.service.Save(ctx, id, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	status := 200
	if id == 0 {
		status = 201
	}
	respondJSON(w, status, result)
}
func (c *PromocionController) Estado(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	id, err := parseID(r)
	if err != nil || id <= 0 {
		respondError(w, 400, "id invalido")
		return
	}
	var req struct {
		Activo *bool `json:"activo"`
	}
	if !decodePromocionBody(w, r, &req) {
		return
	}
	if req.Activo == nil {
		respondJSON(w, 422, map[string]any{"errores": validator.ValidationErrors{{Field: "activo", Message: "es obligatorio"}}})
		return
	}
	result, err := c.service.SetActivo(ctx, id, *req.Activo)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, 200, result)
}
func (c *PromocionController) Cotizar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req models.PedidoCreateRequest
	if !decodePromocionBody(w, r, &req) {
		return
	}
	result, err := c.service.Cotizar(ctx, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, 200, result)
}
