package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ESJ0/selava-backend/internal/service"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	var req loginRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email y contraseña son requeridos")
		return
	}

	token, err := ac.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrCredencialesInvalidas) {
			respondError(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	respondJSON(w, http.StatusOK, loginResponse{Token: token})
}
