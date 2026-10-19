package routes

import (
	"context"
	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type promoRouteRepo struct {
	service.PromocionRepository
	request *models.PedidoCreateRequest
}

func (r *promoRouteRepo) List(context.Context) ([]models.Promocion, error) {
	return []models.Promocion{}, nil
}
func (r *promoRouteRepo) Save(_ context.Context, _ int, p *models.PromocionRequest) (*models.Promocion, error) {
	return &models.Promocion{ID: 1, PromocionRequest: *p}, nil
}
func (r *promoRouteRepo) Cotizar(_ context.Context, p *models.PedidoCreateRequest) (*models.PromocionCotizacion, error) {
	r.request = p
	return &models.PromocionCotizacion{Total: 105, Grupos: []models.PromocionCotizacionGrupo{}}, nil
}
func TestSEL109RolesYContratoCotizacion(t *testing.T) {
	const secret = "sel109-test-only"
	repo := &promoRouteRepo{}
	router := chi.NewRouter()
	router.Route("/api/pedidos", func(r chi.Router) {
		r.Get("/{pedidoID}", func(w http.ResponseWriter, r *http.Request) {
			t.Error("la cotizacion fue confundida con detalle")
			w.WriteHeader(500)
		})
	})
	RegisterPromociones(router, controller.NewPromocionController(service.NewPromocionService(repo)), secret)
	body := `{"nombre":"Dos por cuarenta","tipo_prenda_id":1,"modalidad_cobro":"UNIDAD","medida_requerida":2,"precio_promocional":40,"fecha_inicio":"2026-01-01T00:00:00Z","fecha_fin":"2027-01-01T00:00:00Z","activo":true}`
	for _, role := range []int{1, 2, 3, 0, -1} {
		token := ""
		if role > 0 {
			token, _ = auth.GenerateToken(secret, 1, role)
		} else if role == -1 {
			token, _ = auth.GenerateClienteToken(secret, 1)
		}
		for _, tc := range []struct {
			method, path, body string
			admin, recep       int
		}{{"GET", "/api/promociones/", "", 200, 200}, {"POST", "/api/promociones/", body, 201, 403}, {"POST", "/api/pedidos/cotizacion", `{"total":0,"subtotal_total":0.01,"prendas":[{"tipo_prenda_id":1,"cantidad":5,"precio_aplicado":0.01,"servicios":[{"servicio_id":1}]}],"promociones":[{"tipo_prenda_id":1,"promocion_id":8,"precio_promocional":0.01,"aplicaciones":99}]}`, 200, 200}} {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			expected := 403
			if role == 1 {
				expected = tc.admin
			} else if role == 2 {
				expected = tc.recep
			} else if role <= 0 {
				expected = 401
			}
			if res.Code != expected {
				t.Fatalf("role=%d %s %s status=%d expected=%d body=%s", role, tc.method, tc.path, res.Code, expected, res.Body.String())
			}
		}
	}
	if repo.request == nil || len(repo.request.Promociones) != 1 || repo.request.Promociones[0].PromocionID != 8 || repo.request.Prendas[0].Cantidad != 5 {
		t.Fatal("contrato de seleccion incorrecto")
	}
	// La estructura de request no tiene campos monetarios ni aplicaciones.
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos/cotizacion", strings.NewReader(`{"prendas":[]}`))
	token, _ := auth.GenerateToken(secret, 1, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 422 {
		t.Fatal(res.Code, res.Body.String())
	}
}
