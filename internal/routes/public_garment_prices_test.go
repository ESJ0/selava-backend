package routes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type publicGarmentRepo struct {
	catalog []models.TipoPrendaPublico
	err     error
}

func (r publicGarmentRepo) ListActiveWithPrice(context.Context) ([]models.TipoPrendaPublico, error) {
	return r.catalog, r.err
}
func garmentPricesRouter(repo publicGarmentRepo) http.Handler {
	tc := controller.NewTipoPrendaController(nil, service.NewTipoPrendaPublicoService(repo))
	return NewRouter(nil, nil, nil, nil, tc, nil, nil, nil, nil, nil, nil, "test-secret", "http://localhost:5180")
}
func TestPublicGarmentPricesReadOnlyNoJWT(t *testing.T) {
	router := garmentPricesRouter(publicGarmentRepo{catalog: []models.TipoPrendaPublico{{ID: 1, Nombre: "Camisa", Precio: 25.50, ModalidadCobro: models.ModalidadCobroUnidad}}})
	req := httptest.NewRequest(http.MethodGet, "/api/public/tipos-prenda", nil)
	req.Header.Set("Origin", "http://localhost:5180")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	var catalog []map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if res.Code != 200 || len(catalog) != 1 || len(catalog[0]) != 5 || catalog[0]["precio"] != 25.50 || catalog[0]["modalidad_cobro"] != "UNIDAD" {
		t.Fatal(res.Code, res.Body.String())
	}
	if res.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5180" {
		t.Fatal(res.Header())
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		res = httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest(method, "/api/public/tipos-prenda", nil))
		if res.Code != 405 {
			t.Fatal(method, res.Code)
		}
	}
	res = httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/tipos-prenda/", nil))
	if res.Code != 401 {
		t.Fatal("administracion desprotegida", res.Code)
	}
}
func TestPublicGarmentPricesEmptyAndError(t *testing.T) {
	res := httptest.NewRecorder()
	garmentPricesRouter(publicGarmentRepo{}).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/tipos-prenda", nil))
	if res.Code != 200 || strings.TrimSpace(res.Body.String()) != "[]" {
		t.Fatal(res.Code, res.Body.String())
	}
	res = httptest.NewRecorder()
	garmentPricesRouter(publicGarmentRepo{err: errors.New("sensitive database detail")}).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/public/tipos-prenda", nil))
	if res.Code != 500 || strings.Contains(res.Body.String(), "sensitive") {
		t.Fatal(res.Code, res.Body.String())
	}
}
