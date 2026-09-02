package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

const pedidoTestSecret = "secret"

type fakePedidoRepository struct {
	nextPedidoID int
	nextPrendaID int
	err          error
}

func newPedidoControllerForTest(err error) *PedidoController {
	repo := &fakePedidoRepository{nextPedidoID: 1, nextPrendaID: 1, err: err}
	return NewPedidoController(servicelayer.NewPedidoService(repo))
}

func (r *fakePedidoRepository) Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error) {
	if r.err != nil {
		return nil, r.err
	}

	now := time.Now()
	pedido := models.Pedido{
		ID:            r.nextPedidoID,
		ClienteID:     req.ClienteID,
		UsuarioID:     usuarioID,
		Observaciones: req.Observaciones,
		Activo:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	r.nextPedidoID++

	prendas := make([]models.Prenda, 0, len(req.Prendas))
	for _, p := range req.Prendas {
		prendas = append(prendas, models.Prenda{
			ID:           r.nextPrendaID,
			PedidoID:     pedido.ID,
			TipoPrendaID: p.TipoPrendaID,
			Cantidad:     p.Cantidad,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		r.nextPrendaID++
	}

	return &models.PedidoConPrendas{Pedido: pedido, Prendas: prendas}, nil
}

func (r *fakePedidoRepository) UpdateEstado(_ context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error) {
	if r.err != nil {
		return nil, r.err
	}
	now := time.Now()
	return &models.PedidoEstadoHistorial{
		ID:            1,
		PedidoID:      pedidoID,
		EstadoID:      req.EstadoID,
		UsuarioID:     usuarioID,
		FechaCambio:   now,
		Observaciones: req.Observaciones,
		CreatedAt:     now,
	}, nil
}

func (r *fakePedidoRepository) GetHistorialEstados(_ context.Context, pedidoID int) ([]models.PedidoEstadoHistorial, error) {
	if r.err != nil {
		return nil, r.err
	}
	now := time.Now()
	return []models.PedidoEstadoHistorial{
		{ID: 1, PedidoID: pedidoID, EstadoID: 1, UsuarioID: 7, FechaCambio: now.Add(-time.Hour), CreatedAt: now.Add(-time.Hour)},
		{ID: 2, PedidoID: pedidoID, EstadoID: 2, UsuarioID: 8, FechaCambio: now, CreatedAt: now},
	}, nil
}

func (r *fakePedidoRepository) GetDetalle(_ context.Context, pedidoID int) (*models.PedidoDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &models.PedidoDetalle{
		Pedido:       models.Pedido{ID: pedidoID},
		Cliente:      models.Cliente{ID: 1, Nombre: "Maria"},
		Prendas:      []models.Prenda{{ID: 2, PedidoID: pedidoID}},
		EstadoActual: models.EstadoPedido{ID: 1, Nombre: "Recibido"},
		Pagos:        []models.PagoDetalle{{Pago: models.Pago{ID: 3, PedidoID: pedidoID, Monto: 25}}},
	}, nil
}

func (r *fakePedidoRepository) Cancelar(_ context.Context, pedidoID, usuarioID int) (*models.Pedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &models.Pedido{ID: pedidoID, UsuarioID: usuarioID, EstadoActualID: 2, Activo: true}, nil
}

// authenticatedRequest arma un POST /api/pedidos con un JWT valido, para
// pasar por el mismo camino que usa el middleware real en produccion en
// vez de forjar el contexto directamente (claimsContextKey es privado).
func authenticatedPedidoRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(pedidoTestSecret, 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func servePedidoCrear(controller *PedidoController, req *http.Request) *httptest.ResponseRecorder {
	mw := middleware.NewAuthMiddleware(pedidoTestSecret)
	res := httptest.NewRecorder()
	mw.Authenticate(http.HandlerFunc(controller.Crear)).ServeHTTP(res, req)
	return res
}

func authenticatedPedidoEstadoRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(pedidoTestSecret, 7, middleware.RolOperario)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/pedidos/3/estado", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func servePedidoActualizarEstado(controller *PedidoController, req *http.Request) *httptest.ResponseRecorder {
	mw := middleware.NewAuthMiddleware(pedidoTestSecret)
	router := chi.NewRouter()
	router.With(mw.Authenticate).Put("/api/pedidos/{pedidoID}/estado", controller.ActualizarEstado)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func authenticatedPedidoHistorialRequest(t *testing.T) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(pedidoTestSecret, 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/pedidos/3/historial-estados", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func servePedidoHistorial(controller *PedidoController, req *http.Request) *httptest.ResponseRecorder {
	mw := middleware.NewAuthMiddleware(pedidoTestSecret)
	router := chi.NewRouter()
	router.With(mw.Authenticate).Get("/api/pedidos/{pedidoID}/historial-estados", controller.ObtenerHistorialEstados)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func authenticatedPedidoDetalleRequest(t *testing.T) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(pedidoTestSecret, 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/pedidos/3", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func servePedidoDetalle(controller *PedidoController, req *http.Request) *httptest.ResponseRecorder {
	mw := middleware.NewAuthMiddleware(pedidoTestSecret)
	router := chi.NewRouter()
	router.With(mw.Authenticate).Get("/api/pedidos/{pedidoID}", controller.ObtenerDetalle)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func authenticatedPedidoCancelarRequest(t *testing.T) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(pedidoTestSecret, 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/pedidos/3/cancelar", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func servePedidoCancelar(controller *PedidoController, req *http.Request) *httptest.ResponseRecorder {
	mw := middleware.NewAuthMiddleware(pedidoTestSecret)
	router := chi.NewRouter()
	router.With(mw.Authenticate).Put("/api/pedidos/{pedidoID}/cancelar", controller.Cancelar)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func TestPedidoControllerCancelarReturnsOK(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	res := servePedidoCancelar(controller, authenticatedPedidoCancelarRequest(t))

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
	}
	var body models.Pedido
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.ID != 3 || body.UsuarioID != 7 || body.EstadoActualID != 2 || !body.Activo {
		t.Fatalf("unexpected pedido cancelado: %+v", body)
	}
}

func TestPedidoControllerCancelarRejectsInvalidID(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoCancelarRequest(t)
	req.URL.Path = "/api/pedidos/no-valido/cancelar"

	res := servePedidoCancelar(controller, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, res.Code, res.Body.String())
	}
}

func TestPedidoControllerCancelarReturnsConflictWhenNotCancelable(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrPedidoNoCancelable)
	res := servePedidoCancelar(controller, authenticatedPedidoCancelarRequest(t))

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestPedidoControllerObtenerDetalleReturnsOK(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	res := servePedidoDetalle(controller, authenticatedPedidoDetalleRequest(t))

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
	}
	var body models.PedidoDetalle
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.ID != 3 || body.Cliente.ID != 1 || len(body.Prendas) != 1 || body.EstadoActual.Nombre != "Recibido" || len(body.Pagos) != 1 {
		t.Fatalf("unexpected detalle: %+v", body)
	}
}

func TestPedidoControllerObtenerDetalleRejectsInvalidID(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoDetalleRequest(t)
	req.URL.Path = "/api/pedidos/no-valido"

	res := servePedidoDetalle(controller, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, res.Code, res.Body.String())
	}
}

func TestPedidoControllerObtenerDetalleReturnsNotFound(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrPedidoNoEncontrado)
	res := servePedidoDetalle(controller, authenticatedPedidoDetalleRequest(t))

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestPedidoControllerObtenerHistorialEstadosReturnsOK(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	res := servePedidoHistorial(controller, authenticatedPedidoHistorialRequest(t))

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
	}
	var body []models.PedidoEstadoHistorial
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 2 || body[0].ID != 1 || body[1].ID != 2 {
		t.Fatalf("unexpected historial: %+v", body)
	}
}

func TestPedidoControllerObtenerHistorialEstadosRejectsInvalidID(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoHistorialRequest(t)
	req.URL.Path = "/api/pedidos/no-valido/historial-estados"

	res := servePedidoHistorial(controller, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, res.Code, res.Body.String())
	}
}

func TestPedidoControllerObtenerHistorialEstadosReturnsNotFound(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrPedidoNoEncontrado)
	res := servePedidoHistorial(controller, authenticatedPedidoHistorialRequest(t))

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestPedidoControllerActualizarEstadoReturnsOK(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoEstadoRequest(t, `{"estado_id":2,"observaciones":"En lavado"}`)

	res := servePedidoActualizarEstado(controller, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
	}
	var body models.PedidoEstadoHistorial
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.PedidoID != 3 || body.EstadoID != 2 || body.UsuarioID != 7 {
		t.Fatalf("unexpected historial: %+v", body)
	}
}

func TestPedidoControllerActualizarEstadoRejectsEstadoInvalido(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoEstadoRequest(t, `{"estado_id":0}`)

	res := servePedidoActualizarEstado(controller, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}
}

func TestPedidoControllerActualizarEstadoReturnsNotFound(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrPedidoNoEncontrado)
	req := authenticatedPedidoEstadoRequest(t, `{"estado_id":2}`)

	res := servePedidoActualizarEstado(controller, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestPedidoControllerActualizarEstadoReturnsConflictOnRetroceso(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrPedidoEstadoRetrocedido)
	res := servePedidoActualizarEstado(controller, authenticatedPedidoEstadoRequest(t, `{"estado_id":1}`))

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestPedidoControllerActualizarEstadoReturnsConflictWhenEntregado(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrPedidoEstadoFinalizado)
	res := servePedidoActualizarEstado(controller, authenticatedPedidoEstadoRequest(t, `{"estado_id":4}`))

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestPedidoControllerCrearReturnsCreated(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoRequest(t, `{
		"cliente_id": 1,
		"prendas": [
			{"tipo_prenda_id": 2, "cantidad": 3, "servicios":[{"servicio_id":1}]},
			{"tipo_prenda_id": 5, "cantidad": 1, "servicios":[{"servicio_id":2}]}
		]
	}`)

	res := servePedidoCrear(controller, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}

	var body models.PedidoConPrendas
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.ClienteID != 1 || body.UsuarioID != 7 {
		t.Fatalf("unexpected pedido: %+v", body)
	}
	if len(body.Prendas) != 2 {
		t.Fatalf("expected 2 prendas in response, got %d", len(body.Prendas))
	}
}

func TestPedidoControllerCrearRejectsSinPrendas(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoRequest(t, `{"cliente_id": 1, "prendas": []}`)

	res := servePedidoCrear(controller, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	errores, ok := body["errores"].([]any)
	if !ok || len(errores) == 0 {
		t.Fatalf("expected non-empty errores array, got %+v", body)
	}
}

func TestPedidoControllerCrearRejectsCantidadInvalida(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := authenticatedPedidoRequest(t, `{
		"cliente_id": 1,
		"prendas": [{"tipo_prenda_id": 2, "cantidad": 0, "servicios":[{"servicio_id":1}]}]
	}`)

	res := servePedidoCrear(controller, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}
}

func TestPedidoControllerCrearReturnsNotFoundWhenClienteMissing(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrClienteNoEncontrado)
	req := authenticatedPedidoRequest(t, `{
		"cliente_id": 999,
		"prendas": [{"tipo_prenda_id": 2, "cantidad": 1, "servicios":[{"servicio_id":1}]}]
	}`)

	res := servePedidoCrear(controller, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestPedidoControllerCrearReturnsNotFoundWhenTipoPrendaMissing(t *testing.T) {
	controller := newPedidoControllerForTest(repository.ErrTipoPrendaNoEncontrado)
	req := authenticatedPedidoRequest(t, `{
		"cliente_id": 1,
		"prendas": [{"tipo_prenda_id": 999, "cantidad": 1, "servicios":[{"servicio_id":1}]}]
	}`)

	res := servePedidoCrear(controller, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestPedidoControllerCrearRejectsMissingToken(t *testing.T) {
	controller := newPedidoControllerForTest(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos", strings.NewReader(`{
		"cliente_id": 1,
		"prendas": [{"tipo_prenda_id": 2, "cantidad": 1}]
	}`))

	res := servePedidoCrear(controller, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}
