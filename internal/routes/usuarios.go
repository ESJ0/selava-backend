package routes

import (
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterUsuarios(router chi.Router, c *controller.UsuarioController, jwtSecret string) {
	if c == nil {
		return
	}
	auth := middleware.NewAuthMiddleware(jwtSecret)
	router.Route("/api/usuarios", func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Use(auth.RequireRoles(middleware.RolAdministrador))
		r.Get("/", c.Listar)
		r.Post("/", c.Crear)
		r.Put("/{usuarioID}", c.Actualizar)
		r.Patch("/{usuarioID}/estado", c.Actualizar)
	})
}
