package routes

import (
	"net/http"
	"time"

	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(clienteController *controller.ClienteController) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares livianos: Recoverer evita que un panic tumbe el servidor
	// mientras se está atendiendo a un cliente en caja.
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(5 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/clientes", func(r chi.Router) {
		r.Post("/", clienteController.Crear)
		r.Get("/", clienteController.Listar)
		r.Get("/{id}", clienteController.Obtener)
		r.Put("/{id}", clienteController.Actualizar)
		r.Delete("/{id}", clienteController.Eliminar)
	})

	return r
}