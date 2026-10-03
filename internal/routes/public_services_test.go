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
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

type publicServicesRepo struct {
	service.ServicioRepository
	catalog []models.ServicioPublico
	err     error
}

func (r publicServicesRepo) ListActive(context.Context) ([]models.ServicioPublico, error) {
	return r.catalog, r.err
}

func (r publicServicesRepo) List(context.Context) ([]models.Servicio, error) {
	return []models.Servicio{}, nil
}

func publicServicesRouter(repo publicServicesRepo) http.Handler {
	sc := controller.NewServicioController(service.NewServicioService(repo))
	return NewRouter(nil, sc, nil, nil, nil, nil, nil, nil, nil, nil, nil, "catalog-test-secret", "http://localhost:5174")
}

func TestPublicServicesWithoutJWT(t *testing.T) {
	description, hours := "Cuidado de prendas delicadas", 24
	router := publicServicesRouter(publicServicesRepo{catalog: []models.ServicioPublico{
		{ID: 3, Nombre: "Lavado", Descripcion: &description, PrecioBase: 37.25, TiempoEstimadoHoras: &hours},
	}})
	req := httptest.NewRequest(http.MethodGet, "/api/public/servicios", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.HasPrefix(res.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("status=%d headers=%v body=%s", res.Code, res.Header(), res.Body.String())
	}
	var catalog []map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 1 || catalog[0]["precio_base"] != 37.25 || catalog[0]["nombre"] != "Lavado" {
		t.Fatalf("catalogo inesperado: %v", catalog)
	}
	if len(catalog[0]) != 5 || catalog[0]["descripcion"] != description || catalog[0]["tiempo_estimado_horas"] != float64(hours) {
		t.Fatalf("campos publicos inesperados: %v", catalog[0])
	}
	for _, field := range []string{"activo", "created_at", "updated_at"} {
		if _, exists := catalog[0][field]; exists {
			t.Fatalf("campo interno expuesto: %s", field)
		}
	}
}

func TestPublicServicesEmptyCatalog(t *testing.T) {
	for _, catalog := range [][]models.ServicioPublico{nil, {}} {
		res := httptest.NewRecorder()
		publicServicesRouter(publicServicesRepo{catalog: catalog}).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/servicios", nil))
		if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != "[]" {
			t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
		}
	}
}

func TestPublicServicesDatabaseErrorIsSanitized(t *testing.T) {
	res := httptest.NewRecorder()
	publicServicesRouter(publicServicesRepo{err: errors.New("sensitive database detail")}).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/servicios", nil))
	if res.Code != http.StatusInternalServerError || strings.Contains(res.Body.String(), "sensitive") {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestPublicServicesReadOnlyAndAdminStillProtected(t *testing.T) {
	router := publicServicesRouter(publicServicesRepo{})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest(method, "/api/public/servicios", nil))
		if res.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s publico status=%d", method, res.Code)
		}
	}
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/servicios/"}, {http.MethodPost, "/api/servicios/"},
		{http.MethodGet, "/api/servicios/1"}, {http.MethodPut, "/api/servicios/1"}, {http.MethodDelete, "/api/servicios/1"},
	} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest(tc.method, tc.path, nil))
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, res.Code)
		}
	}
	for _, tc := range []struct{ role, want int }{
		{middleware.RolAdministrador, http.StatusOK}, {middleware.RolRecepcionista, http.StatusOK}, {middleware.RolOperario, http.StatusForbidden},
	} {
		token, err := auth.GenerateToken("catalog-test-secret", 1, tc.role)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/servicios/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != tc.want {
			t.Fatalf("rol=%d status=%d esperado=%d", tc.role, res.Code, tc.want)
		}
	}
}

func TestPublicServicesCORS(t *testing.T) {
	router := publicServicesRouter(publicServicesRepo{})
	for _, tc := range []struct {
		method, origin string
		status         int
		allowed        bool
	}{
		{http.MethodGet, "http://localhost:5174", http.StatusOK, true},
		{http.MethodOptions, "http://localhost:5174", http.StatusNoContent, true},
		{http.MethodGet, "https://untrusted.example", http.StatusOK, false},
		{http.MethodOptions, "https://untrusted.example", http.StatusForbidden, false},
	} {
		req := httptest.NewRequest(tc.method, "/api/public/servicios", nil)
		req.Header.Set("Origin", tc.origin)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != tc.status {
			t.Fatalf("%s %s status=%d", tc.method, tc.origin, res.Code)
		}
		want := ""
		if tc.allowed {
			want = tc.origin
		}
		if res.Header().Get("Access-Control-Allow-Origin") != want {
			t.Fatalf("CORS inesperado: %v", res.Header())
		}
	}
}
