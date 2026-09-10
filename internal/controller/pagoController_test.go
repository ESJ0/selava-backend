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

type fakePagoControllerRepository struct{ err error }

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
