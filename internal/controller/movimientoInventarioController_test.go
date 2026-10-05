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
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
)

type fakeMovimientoControllerRepository struct {
	usuarioID int
	request   models.MovimientoInventarioCreateRequest
}

func (r *fakeMovimientoControllerRepository) Create(ctx context.Context, req *models.MovimientoInventarioCreateRequest, usuarioID int) (*models.MovimientoInventario, error) {
	r.usuarioID = usuarioID
	r.request = *req
	return &models.MovimientoInventario{
		ID: 1, InsumoID: req.InsumoID, UsuarioID: usuarioID,
		TipoMovimiento: req.TipoMovimiento, Cantidad: req.Cantidad,
		Motivo: req.Motivo,
	}, nil
}

func TestMovimientoInventarioControllerRegistraConUsuarioDelToken(t *testing.T) {
	repo := &fakeMovimientoControllerRepository{}
	controller := NewMovimientoInventarioController(servicelayer.NewMovimientoInventarioService(repo))
	authMiddleware := middleware.NewAuthMiddleware("test-secret")
	token, err := auth.GenerateToken("test-secret", 7, middleware.RolOperario)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/movimientos-inventario", strings.NewReader(`{"insumo_id":4,"tipo_movimiento":"salida","cantidad":2}`))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	authMiddleware.Authenticate(http.HandlerFunc(controller.Registrar)).ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
	var body models.MovimientoInventario
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.InsumoID != 4 || body.UsuarioID != 7 || repo.usuarioID != 7 {
		t.Fatalf("unexpected movement/user: body=%+v repo=%d", body, repo.usuarioID)
	}
}

func TestMovimientoInventarioControllerRechazaSinToken(t *testing.T) {
	controller := NewMovimientoInventarioController(servicelayer.NewMovimientoInventarioService(&fakeMovimientoControllerRepository{}))
	res := httptest.NewRecorder()

	controller.Registrar(res, httptest.NewRequest(http.MethodPost, "/api/movimientos-inventario", strings.NewReader(`{"insumo_id":4,"tipo_movimiento":"entrada","cantidad":1}`)))

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}
