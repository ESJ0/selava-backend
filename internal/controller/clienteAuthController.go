package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	ObtenerPerfil(context.Context, int) (*models.ClienteResponse, error)
	ActualizarPerfil(context.Context, int, *models.ClientePerfilUpdateRequest) (*models.ClienteResponse, error)
	CambiarPassword(context.Context, int, *models.ClientePasswordUpdateRequest) error
	GuardarFotoPerfil(context.Context, int, []byte, string) error
	ObtenerFotoPerfil(context.Context, int) (*models.ClienteFotoPerfil, error)
	EliminarFotoPerfil(context.Context, int) error
}

const maxFotoPerfilBytes = 2 << 20

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
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	claims, ok := middleware.ClienteClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return
	}
	cliente, err := c.service.ObtenerPerfil(ctx, claims.ClienteID)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, cliente)
}

func (c *ClienteAuthController) ActualizarPerfil(w http.ResponseWriter, r *http.Request) {
	clienteID, ok := clienteIDFromRequest(w, r)
	if !ok {
		return
	}
	var req models.ClientePerfilUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	perfil, err := c.service.ActualizarPerfil(ctx, clienteID, &req)
	if err != nil {
		c.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, perfil)
}

func (c *ClienteAuthController) CambiarPassword(w http.ResponseWriter, r *http.Request) {
	clienteID, ok := clienteIDFromRequest(w, r)
	if !ok {
		return
	}
	var req models.ClientePasswordUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	if err := c.service.CambiarPassword(ctx, clienteID, &req); err != nil {
		c.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *ClienteAuthController) GuardarFotoPerfil(w http.ResponseWriter, r *http.Request) {
	clienteID, ok := clienteIDFromRequest(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFotoPerfilBytes+1024)
	if err := r.ParseMultipartForm(maxFotoPerfilBytes); err != nil {
		respondError(w, http.StatusBadRequest, "foto invalida o demasiado grande")
		return
	}
	file, _, err := r.FormFile("foto")
	if err != nil {
		respondError(w, http.StatusBadRequest, "foto es requerida")
		return
	}
	defer file.Close()
	contenido, err := io.ReadAll(io.LimitReader(file, maxFotoPerfilBytes+1))
	if err != nil || len(contenido) == 0 || len(contenido) > maxFotoPerfilBytes {
		respondError(w, http.StatusBadRequest, "la foto no puede superar 2 MB")
		return
	}
	tipo := http.DetectContentType(contenido)
	if tipo != "image/jpeg" && tipo != "image/png" && tipo != "image/webp" {
		respondError(w, http.StatusUnprocessableEntity, "la foto debe ser JPEG, PNG o WebP")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	if err = c.service.GuardarFotoPerfil(ctx, clienteID, contenido, tipo); err != nil {
		c.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *ClienteAuthController) FotoPerfil(w http.ResponseWriter, r *http.Request) {
	clienteID, ok := clienteIDFromRequest(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	foto, err := c.service.ObtenerFotoPerfil(ctx, clienteID)
	if err != nil {
		if errors.Is(err, repository.ErrCuentaClienteNoEncontrada) {
			respondError(w, http.StatusNotFound, "foto de perfil no encontrada")
			return
		}
		c.handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", foto.Tipo)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(foto.Contenido)
}

func (c *ClienteAuthController) EliminarFotoPerfil(w http.ResponseWriter, r *http.Request) {
	clienteID, ok := clienteIDFromRequest(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	if err := c.service.EliminarFotoPerfil(ctx, clienteID); err != nil {
		c.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func clienteIDFromRequest(w http.ResponseWriter, r *http.Request) (int, bool) {
	claims, ok := middleware.ClienteClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "token de autenticacion requerido")
		return 0, false
	}
	return claims.ClienteID, true
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
	case errors.Is(err, service.ErrPasswordActualIncorrecta):
		respondError(w, http.StatusUnauthorized, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
