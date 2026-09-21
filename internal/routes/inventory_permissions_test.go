package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

const inventoryPermissionsSecret = "inventory-permissions-test-secret"

type inventoryRouteRepository struct{}

func (inventoryRouteRepository) Create(_ context.Context, req *models.InsumoCreateRequest) (*models.Insumo, error) {
	return &models.Insumo{ID: 1, Nombre: req.Nombre, UnidadMedida: req.UnidadMedida, StockActual: req.StockActual, StockMinimo: req.StockMinimo, Activo: true}, nil
}

func (inventoryRouteRepository) GetByID(_ context.Context, id int) (*models.Insumo, error) {
	return &models.Insumo{ID: id, Nombre: "Detergente", UnidadMedida: "L", StockActual: 10, StockMinimo: 5, Activo: true}, nil
}

func (inventoryRouteRepository) List(_ context.Context) ([]models.Insumo, error) {
	return []models.Insumo{{ID: 1, Nombre: "Detergente", UnidadMedida: "L", StockActual: 10, StockMinimo: 5, Activo: true}}, nil
}

func (inventoryRouteRepository) ListLowStock(_ context.Context) ([]models.Insumo, error) {
	return []models.Insumo{}, nil
}

func (inventoryRouteRepository) Update(_ context.Context, id int, req *models.InsumoUpdateRequest) (*models.Insumo, error) {
	return &models.Insumo{ID: id, Nombre: "Detergente", UnidadMedida: "L", StockActual: 10, StockMinimo: 5, Activo: true}, nil
}

func (inventoryRouteRepository) Delete(_ context.Context, _ int) error { return nil }

type movementRouteRepository struct{}

func (movementRouteRepository) Create(_ context.Context, req *models.MovimientoInventarioCreateRequest, userID int) (*models.MovimientoInventario, error) {
	return &models.MovimientoInventario{ID: 1, InsumoID: req.InsumoID, UsuarioID: userID, TipoMovimiento: req.TipoMovimiento, Cantidad: req.Cantidad, FechaMovimiento: time.Now()}, nil
}

func inventoryPermissionsRouter() http.Handler {
	inputController := controller.NewInsumoController(service.NewInsumoService(inventoryRouteRepository{}))
	movementController := controller.NewMovimientoInventarioController(service.NewMovimientoInventarioService(movementRouteRepository{}))
	return NewRouter(nil, nil, inputController, movementController, nil, nil, nil, nil, nil, nil, nil, inventoryPermissionsSecret, "")
}

func inventoryRoleRequest(t *testing.T, method, target, body string, role int) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(inventoryPermissionsSecret, 7, role)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestOperarioPuedeLeerInsumosYRegistrarMovimientos(t *testing.T) {
	router := inventoryPermissionsRouter()
	for _, target := range []string{"/api/insumos/", "/api/insumos/1", "/api/insumos/alertas/stock-minimo"} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, inventoryRoleRequest(t, http.MethodGet, target, "", middleware.RolOperario))
		if res.Code != http.StatusOK {
			t.Fatalf("GET %s status=%d, esperado=%d: %s", target, res.Code, http.StatusOK, res.Body.String())
		}
	}

	res := httptest.NewRecorder()
	router.ServeHTTP(res, inventoryRoleRequest(t, http.MethodPost, "/api/movimientos-inventario/", `{"insumo_id":1,"tipo_movimiento":"entrada","cantidad":2}`, middleware.RolOperario))
	if res.Code != http.StatusCreated {
		t.Fatalf("movimiento operario status=%d, esperado=%d: %s", res.Code, http.StatusCreated, res.Body.String())
	}
}

func TestOperarioNoPuedeModificarCatalogoDeInsumos(t *testing.T) {
	router := inventoryPermissionsRouter()
	for _, tc := range []struct {
		method string
		target string
		body   string
	}{
		{http.MethodPost, "/api/insumos/", `{"nombre":"Cloro","unidad_medida":"L"}`},
		{http.MethodPut, "/api/insumos/1", `{"stock_minimo":4}`},
		{http.MethodDelete, "/api/insumos/1", ""},
	} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, inventoryRoleRequest(t, tc.method, tc.target, tc.body, middleware.RolOperario))
		if res.Code != http.StatusForbidden {
			t.Fatalf("%s %s status=%d, esperado=%d: %s", tc.method, tc.target, res.Code, http.StatusForbidden, res.Body.String())
		}
	}
}

func TestAdministradorMantienePermisosDeInventario(t *testing.T) {
	router := inventoryPermissionsRouter()
	for _, tc := range []struct {
		method string
		target string
		body   string
		want   int
	}{
		{http.MethodGet, "/api/insumos/", "", http.StatusOK},
		{http.MethodPost, "/api/insumos/", `{"nombre":"Cloro","unidad_medida":"L","stock_actual":0,"stock_minimo":2}`, http.StatusCreated},
		{http.MethodPost, "/api/movimientos-inventario/", `{"insumo_id":1,"tipo_movimiento":"salida","cantidad":1}`, http.StatusCreated},
	} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, inventoryRoleRequest(t, tc.method, tc.target, tc.body, middleware.RolAdministrador))
		if res.Code != tc.want {
			t.Fatalf("%s %s status=%d, esperado=%d: %s", tc.method, tc.target, res.Code, tc.want, res.Body.String())
		}
	}
}
