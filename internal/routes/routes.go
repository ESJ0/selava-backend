package routes

import (
	"net/http"
	"time"

	"github.com/ESJ0/selava-backend/internal/controller"
	authmiddleware "github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(clienteController *controller.ClienteController, servicioController *controller.ServicioController, tipoPrendaController *controller.TipoPrendaController, metodoPagoController *controller.MetodoPagoController, pedidoController *controller.PedidoController, prendaController *controller.PrendaController, authController *controller.AuthController, jwtSecret, allowedOrigins string) *chi.Mux {
	r := chi.NewRouter()
	authMW := authmiddleware.NewAuthMiddleware(jwtSecret)

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(5 * time.Second))
	r.Use(authmiddleware.CORS(allowedOrigins))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Post("/api/auth/login", authController.Login)

	r.Route("/api/clientes", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista))

		r.Post("/", clienteController.Crear)
		r.Get("/", clienteController.Listar)
		r.Get("/{id}", clienteController.Obtener)
		r.Put("/{id}", clienteController.Actualizar)
		r.Delete("/{id}", clienteController.Eliminar)
	})

	r.Route("/api/servicios", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista))

		r.Post("/", servicioController.Crear)
		r.Get("/", servicioController.Listar)
		r.Get("/{id}", servicioController.Obtener)
		r.Put("/{id}", servicioController.Actualizar)
		r.Delete("/{id}", servicioController.Eliminar)
	})

	r.Route("/api/tipos-prenda", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista))

		r.Post("/", tipoPrendaController.Crear)
		r.Get("/", tipoPrendaController.Listar)
		r.Get("/{id}", tipoPrendaController.Obtener)
		r.Put("/{id}", tipoPrendaController.Actualizar)
		r.Delete("/{id}", tipoPrendaController.Eliminar)
	})

	r.Route("/api/metodos-pago", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista))
		r.Post("/", metodoPagoController.Crear)
		r.Get("/", metodoPagoController.Listar)
		r.Get("/{id}", metodoPagoController.Obtener)
		r.Put("/{id}", metodoPagoController.Actualizar)
		r.Delete("/{id}", metodoPagoController.Eliminar)
	})

	r.Route("/api/pedidos", func(r chi.Router) {
		r.Use(authMW.Authenticate)

		r.With(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista)).Post("/", pedidoController.Crear)
		r.Get("/{pedidoID}", pedidoController.ObtenerDetalle)
		r.With(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista)).Put("/{pedidoID}/cancelar", pedidoController.Cancelar)
		r.With(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista)).Post("/{pedidoID}/prendas", prendaController.CrearVarias)
		r.Get("/{pedidoID}/historial-estados", pedidoController.ObtenerHistorialEstados)
		r.With(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista, authmiddleware.RolOperario)).Put("/{pedidoID}/estado", pedidoController.ActualizarEstado)
	})

	r.Route("/api/prendas", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRoles(authmiddleware.RolAdministrador, authmiddleware.RolRecepcionista))

		r.Post("/{prendaID}/servicios", prendaController.AsociarServicio)
		r.Delete("/{prendaID}/servicios/{servicioID}", prendaController.QuitarServicio)
	})

	return r
}
