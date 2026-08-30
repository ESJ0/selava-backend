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
