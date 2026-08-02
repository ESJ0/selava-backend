package routes

import (
	"net/http"
	"time"

	"github.com/ESJ0/selava-backend/internal/controller"
	authmiddleware "github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(clienteController *controller.ClienteController, authController *controller.AuthController, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()
	authMW := authmiddleware.NewAuthMiddleware(jwtSecret)

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(5 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Post("/api/auth/login", authController.Login)

	r.Route("/api/clientes", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolEmpleado))

		r.Post("/", clienteController.Crear)
		r.Get("/", clienteController.Listar)
		r.Get("/{id}", clienteController.Obtener)
		r.Put("/{id}", clienteController.Actualizar)
		r.Delete("/{id}", clienteController.Eliminar)
	})

	return r
}
