package routes

import (
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterPromociones(r chi.Router, c *controller.PromocionController, jwtSecret string) {
	auth := middleware.NewAuthMiddleware(jwtSecret)
	r.Route("/api/promociones", func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Use(auth.RequireRoles(middleware.RolAdministrador, middleware.RolRecepcionista))
		r.Get("/", c.Listar)
		r.With(auth.RequireRoles(middleware.RolAdministrador)).Post("/", c.Crear)
		r.Get("/{id}", c.Obtener)
		r.With(auth.RequireRoles(middleware.RolAdministrador)).Put("/{id}", c.Actualizar)
		r.With(auth.RequireRoles(middleware.RolAdministrador)).Patch("/{id}/estado", c.Estado)
	})
	r.With(auth.Authenticate, auth.RequireRoles(middleware.RolAdministrador, middleware.RolRecepcionista)).Post("/api/pedidos/cotizacion", c.Cotizar)
}
