package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type PasswordResetController struct{ service *service.PasswordResetService }

func NewPasswordResetController(service *service.PasswordResetService) *PasswordResetController {
	return &PasswordResetController{service: service}
}

func (c *PasswordResetController) Solicitar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	if err := c.service.Solicitar(ctx, req.Email); err != nil {
		log.Printf("solicitud de recuperacion fallida: %v", err)
	}
	respondJSON(w, http.StatusAccepted, map[string]string{"mensaje": "Si el correo existe, recibiras un enlace de recuperacion."})
}

func (c *PasswordResetController) Validar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	if err := c.service.Validar(ctx, r.URL.Query().Get("token")); err != nil {
		respondError(w, http.StatusBadRequest, repository.ErrTokenRecuperacionInvalido.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"valido": true})
}

func (c *PasswordResetController) Restablecer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	err := c.service.Restablecer(ctx, req.Token, req.Password)
	var validation validator.ValidationErrors
	switch {
	case errors.As(err, &validation):
		respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": validation})
	case errors.Is(err, repository.ErrTokenRecuperacionInvalido):
		respondError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	default:
		respondJSON(w, http.StatusOK, map[string]string{"mensaje": "Contrasena actualizada correctamente."})
	}
}
