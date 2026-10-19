package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSEL109PreflightEstadoPromocion(t *testing.T) {
	handler := CORS("http://localhost:5195")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	for _, origin := range []string{"http://localhost:5195", "https://otro.test"} {
		req := httptest.NewRequest(http.MethodOptions, "/api/promociones/1/estado", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "PATCH")
		req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if origin == "http://localhost:5195" {
			if res.Code != 204 || !strings.Contains(res.Header().Get("Access-Control-Allow-Methods"), "PATCH") || res.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Fatal(res.Code, res.Header())
			}
		} else if res.Code != 403 || res.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("origen ajeno permitido", res.Code, res.Header())
		}
	}
}
