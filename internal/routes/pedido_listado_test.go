package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

type listadoRouteRepo struct {
	service.PedidoRepository
	filter models.PedidoListFilter
	err    error
}

func (r *listadoRouteRepo) List(_ context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error) {
	r.filter = filter
	return &models.PedidoListResponse{Pedidos: []models.PedidoResumen{}, Pagina: filter.Pagina, Limite: filter.Limite}, r.err
}

func listadoRouter(repo *listadoRouteRepo) http.Handler {
	pc := controller.NewPedidoController(service.NewPedidoService(repo))
	return NewRouter(nil, nil, nil, nil, nil, nil, pc, nil, nil, nil, nil, "listado-test-secret", "http://localhost:5199")
}

func listadoRequest(t *testing.T, target string, role int) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if role != 0 {
		token, err := auth.GenerateToken("listado-test-secret", 1, role)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestPedidoListadoRequiresAuthorizedRole(t *testing.T) {
	for _, tc := range []struct{ role, status int }{{0, 401}, {1, 200}, {2, 200}, {3, 200}, {99, 403}} {
		res := httptest.NewRecorder()
		listadoRouter(&listadoRouteRepo{}).ServeHTTP(res, listadoRequest(t, "/api/pedidos/", tc.role))
		if res.Code != tc.status {
			t.Fatalf("role=%d status=%d body=%s", tc.role, res.Code, res.Body.String())
		}
	}
}

func TestPedidoListadoQueryAndEmptyResponse(t *testing.T) {
	repo := &listadoRouteRepo{}
	res := httptest.NewRecorder()
	listadoRouter(repo).ServeHTTP(res, listadoRequest(t, "/api/pedidos/?q=%20Ana%20&estado_id=3&fecha_desde=2026-10-05&fecha_hasta=2026-10-05&pagina=2&limite=50&orden=antiguos", 1))
	var result models.PedidoListResponse
	if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if res.Code != 200 || result.Pedidos == nil || result.Total != 0 || result.Pagina != 2 || result.Limite != 50 || repo.filter.Buscar != "Ana" || repo.filter.EstadoID != 3 || repo.filter.Orden != "antiguos" {
		t.Fatalf("status=%d body=%s filter=%+v", res.Code, res.Body.String(), repo.filter)
	}
}

func TestPedidoListadoInvalidQueryAndSanitizedError(t *testing.T) {
	res := httptest.NewRecorder()
	listadoRouter(&listadoRouteRepo{}).ServeHTTP(res, listadoRequest(t, "/api/pedidos/?fecha_desde=2026-10-06&fecha_hasta=2026-10-05", 1))
	if res.Code != 422 || !strings.Contains(res.Body.String(), "fecha_hasta") {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	res = httptest.NewRecorder()
	listadoRouter(&listadoRouteRepo{err: errors.New("private SQL failure")}).ServeHTTP(res, listadoRequest(t, "/api/pedidos/", 1))
	if res.Code != 500 || strings.Contains(res.Body.String(), "private") {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}
