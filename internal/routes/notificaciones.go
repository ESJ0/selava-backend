package routes

import (
	"github.com/ESJ0/selava-backend/internal/controller"
	authmiddleware "github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// RegisterNotificacionesCliente mantiene el modulo desacoplado del constructor
// historico del router y permite montarlo tambien en pruebas focalizadas.
func RegisterNotificacionesCliente(router chi.Router, notificacionController *controller.NotificacionController, jwtSecret string) {
	if notificacionController == nil {
		return
	}
	authMW := authmiddleware.NewAuthMiddleware(jwtSecret)
	router.Route("/api/cliente/notificaciones", func(r chi.Router) {
		r.Use(authMW.AuthenticateCliente)
		r.Get("/", notificacionController.ListarPropias)
		r.Patch("/leidas", notificacionController.MarcarTodasLeidas)
		r.Patch("/{notificacionID}/leida", notificacionController.MarcarLeida)
		r.Delete("/{notificacionID}", notificacionController.Eliminar)
	})
}
