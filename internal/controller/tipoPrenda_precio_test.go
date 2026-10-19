package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTipoPrendaPrecioHTTP(t *testing.T) {
	for _, body := range []string{`{"nombre":"Camisa"}`, `{"nombre":"Camisa","precio":null}`, `{"nombre":"Camisa","precio":0}`, `{"nombre":"Camisa","precio":-1}`, `{"nombre":"Camisa","precio":25.001}`} {
		res := httptest.NewRecorder()
		newTipoPrendaControllerForTest().Crear(res, httptest.NewRequest(http.MethodPost, "/api/tipos-prenda/", strings.NewReader(body)))
		var payload struct{ Errores []struct{ Field string } }
		if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if res.Code != 422 || len(payload.Errores) != 1 || payload.Errores[0].Field != "precio" {
			t.Fatalf("%s: %d %s", body, res.Code, res.Body.String())
		}
	}
	tc := newTipoPrendaControllerForTest()
	res := httptest.NewRecorder()
	tc.Crear(res, httptest.NewRequest(http.MethodPost, "/api/tipos-prenda/", strings.NewReader(`{"nombre":"Camisa","precio":25.50}`)))
	if res.Code != 201 || !strings.Contains(res.Body.String(), `"precio":25.5`) {
		t.Fatal(res.Code, res.Body.String())
	}
	res = httptest.NewRecorder()
	tc.Actualizar(res, withIDParam(http.MethodPut, "/api/tipos-prenda/1", `{"precio":30.25}`, 1))
	if res.Code != 200 || !strings.Contains(res.Body.String(), `"precio":30.25`) {
		t.Fatal(res.Code, res.Body.String())
	}
	res = httptest.NewRecorder()
	tc.Obtener(res, withIDParam(http.MethodGet, "/api/tipos-prenda/1", "", 1))
	if res.Code != 200 || !strings.Contains(res.Body.String(), `"precio":30.25`) {
		t.Fatal(res.Code, res.Body.String())
	}
	res = httptest.NewRecorder()
	tc.Listar(res, httptest.NewRequest(http.MethodGet, "/api/tipos-prenda/", nil))
	if res.Code != 200 || !strings.Contains(res.Body.String(), `"precio":30.25`) {
		t.Fatal(res.Code, res.Body.String())
	}
}

func TestServicioNoExigePrecio(t *testing.T) {
	res := httptest.NewRecorder()
	newServicioControllerForTest().Crear(res, httptest.NewRequest(http.MethodPost, "/api/servicios/", strings.NewReader(`{"nombre":"Lavado"}`)))
	if res.Code != 201 {
		t.Fatal(res.Code, res.Body.String())
	}
}
