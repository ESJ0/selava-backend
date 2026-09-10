package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type fakePagoControllerRepository struct {
	err   error
	saldo *models.SaldoPedido
	pagos []models.PagoDetalle
}

func (r *fakePagoControllerRepository) ListByPedido(_ context.Context, _ int) ([]models.PagoDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.pagos == nil {
		return []models.PagoDetalle{}, nil
	}
	return r.pagos, nil
}

func (r *fakePagoControllerRepository) GetSaldo(_ context.Context, _ int) (*models.SaldoPedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.saldo, nil
}

func (r *fakePagoControllerRepository) Create(_ context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &models.PagoDetalle{Pago: models.Pago{ID: 9, PedidoID: pedidoID, MetodoPagoID: req.MetodoPagoID, UsuarioID: usuarioID, Monto: req.Monto}}, nil
}

func pagoRegistrarHandler(t *testing.T, repo *fakePagoControllerRepository) http.Handler {
	t.Helper()
	controller := NewPagoController(servicelayer.NewPagoService(repo))
	authMW := middleware.NewAuthMiddleware("pago-test-secret")
	router := chi.NewRouter()
	router.With(authMW.Authenticate).Post("/api/pedidos/{pedidoID}/pagos", controller.Registrar)
	return router
}

func pagoSaldoHandler(t *testing.T, repo *fakePagoControllerRepository) http.Handler {
	t.Helper()
	controller := NewPagoController(servicelayer.NewPagoService(repo))
	authMW := middleware.NewAuthMiddleware("pago-test-secret")
	router := chi.NewRouter()
	router.With(authMW.Authenticate).Get("/api/pedidos/{pedidoID}/saldo", controller.ObtenerSaldo)
	return router
}

func pagoHistorialHandler(t *testing.T, repo *fakePagoControllerRepository) http.Handler {
	t.Helper()
	controller := NewPagoController(servicelayer.NewPagoService(repo))
	authMW := middleware.NewAuthMiddleware("pago-test-secret")
	router := chi.NewRouter()
	router.With(authMW.Authenticate).Get("/api/pedidos/{pedidoID}/pagos", controller.ObtenerHistorial)
	return router
}

func pagoRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken("pago-test-secret", 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos/4/pagos", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func pagoAuthenticatedRequest(t *testing.T, method, target, body string) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken("pago-test-secret", 7, middleware.RolRecepcionista)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func TestPagoControllerRegistrarReturnsCreated(t *testing.T) {
	res := httptest.NewRecorder()
	pagoRegistrarHandler(t, &fakePagoControllerRepository{}).ServeHTTP(res, pagoRequest(t, `{"metodo_pago_id":2,"monto":25.5}`))
	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var pago models.PagoDetalle
	if err := json.NewDecoder(res.Body).Decode(&pago); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if pago.PedidoID != 4 || pago.UsuarioID != 7 || pago.Monto != 25.5 {
		t.Fatalf("pago inesperado: %+v", pago)
	}
}

func TestPagoControllerRegistrarReturnsConflictWhenMontoExceedsSaldo(t *testing.T) {
	res := httptest.NewRecorder()
	pagoRegistrarHandler(t, &fakePagoControllerRepository{err: repository.ErrPagoExcedeSaldo}).ServeHTTP(res, pagoRequest(t, `{"metodo_pago_id":2,"monto":25.5}`))
	if res.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestPagoControllerRegistrarReturnsUnprocessableEntity(t *testing.T) {
	res := httptest.NewRecorder()
	pagoRegistrarHandler(t, &fakePagoControllerRepository{}).ServeHTTP(res, pagoRequest(t, `{"metodo_pago_id":0,"monto":0}`))
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestPagoControllerObtenerSaldoReturnsOK(t *testing.T) {
	repo := &fakePagoControllerRepository{saldo: &models.SaldoPedido{PedidoID: 4, Total: 80, TotalPagado: 30, SaldoPendiente: 50}}
	res := httptest.NewRecorder()
	pagoSaldoHandler(t, repo).ServeHTTP(res, pagoAuthenticatedRequest(t, http.MethodGet, "/api/pedidos/4/saldo", ""))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var saldo models.SaldoPedido
	if err := json.NewDecoder(res.Body).Decode(&saldo); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if saldo.SaldoPendiente != 50 || saldo.TotalPagado != 30 {
		t.Fatalf("saldo inesperado: %+v", saldo)
	}
}

func TestPagoControllerObtenerSaldoReturnsNotFound(t *testing.T) {
	res := httptest.NewRecorder()
	pagoSaldoHandler(t, &fakePagoControllerRepository{err: repository.ErrPedidoNoEncontrado}).ServeHTTP(res, pagoAuthenticatedRequest(t, http.MethodGet, "/api/pedidos/99/saldo", ""))
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestPagoControllerObtenerHistorialReturnsOK(t *testing.T) {
	repo := &fakePagoControllerRepository{pagos: []models.PagoDetalle{{Pago: models.Pago{ID: 1, PedidoID: 4, Monto: 30}}}}
	res := httptest.NewRecorder()
	pagoHistorialHandler(t, repo).ServeHTTP(res, pagoAuthenticatedRequest(t, http.MethodGet, "/api/pedidos/4/pagos", ""))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var pagos []models.PagoDetalle
	if err := json.NewDecoder(res.Body).Decode(&pagos); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(pagos) != 1 || pagos[0].Monto != 30 {
		t.Fatalf("historial inesperado: %+v", pagos)
	}
}

func TestPagoControllerObtenerHistorialReturnsEmptyArray(t *testing.T) {
	res := httptest.NewRecorder()
	pagoHistorialHandler(t, &fakePagoControllerRepository{}).ServeHTTP(res, pagoAuthenticatedRequest(t, http.MethodGet, "/api/pedidos/4/pagos", ""))
	if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != "[]" {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}
