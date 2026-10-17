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

type clienteAuthService interface {
	Registrar(context.Context, *models.ClienteRegistroRequest) (*models.ClienteRegistrado, error)
	Login(context.Context, *models.ClienteLoginRequest) (string, error)
}

type ClienteAuthController struct {
	service clienteAuthService
}

func NewClienteAuthController(service clienteAuthService) *ClienteAuthController {
	return &ClienteAuthController{service: service}
}

func (c *ClienteAuthController) Registrar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.ClienteRegistroRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}

	registrado, err := c.service.Registrar(ctx, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, registrado.ToResponse())
}

func (c *ClienteAuthController) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req models.ClienteLoginRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}

	token, err := c.service.Login(ctx, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, loginResponse{Token: token})
}

// Me es una ruta privada minima que permite al frontend comprobar la sesion y
// demuestra que la identidad se obtiene del JWT, nunca de un ID enviado por el
// cliente en URL o body.
func (c *ClienteAuthController) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClienteClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return
	}
	respondJSON(w, http.StatusOK, map[string]int{"cliente_id": claims.ClienteID})
}

func (c *ClienteAuthController) handleError(w http.ResponseWriter, err error) {
	var verrs validator.ValidationErrors
	switch {
	case errors.As(err, &verrs):
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": verrs})
	case errors.Is(err, repository.ErrEmailCuentaClienteEnUso):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrCredencialesClienteInvalidas):
		respondError(w, http.StatusUnauthorized, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
