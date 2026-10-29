package routes

import (
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/go-chi/chi/v5"
)

func RegisterPasswordReset(router chi.Router, c *controller.PasswordResetController) {
	if c == nil {
		return
	}
	router.Post("/api/auth/password/forgot", c.Solicitar)
	router.Get("/api/auth/password/reset/validate", c.Validar)
	router.Post("/api/auth/password/reset", c.Restablecer)
}
