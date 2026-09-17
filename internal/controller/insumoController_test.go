package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
)

type fakeInsumoControllerRepository struct {
	insumos map[int]models.Insumo
	nextID  int
}

func newInsumoControllerForTest(insumos ...models.Insumo) *InsumoController {
	repo := &fakeInsumoControllerRepository{insumos: make(map[int]models.Insumo), nextID: 1}
	for _, insumo := range insumos {
		repo.insumos[insumo.ID] = insumo
		if insumo.ID >= repo.nextID {
			repo.nextID = insumo.ID + 1
		}
	}
	return NewInsumoController(servicelayer.NewInsumoService(repo))
}

func insumoControllerFixture(id int, nombre string) models.Insumo {
	now := time.Now()
	return models.Insumo{
		ID: id, Nombre: nombre, UnidadMedida: "litros", StockActual: 10,
		StockMinimo: 2, Activo: true, CreatedAt: now, UpdatedAt: now,
	}
}

func (r *fakeInsumoControllerRepository) Create(ctx context.Context, req *models.InsumoCreateRequest) (*models.Insumo, error) {
	for _, insumo := range r.insumos {
		if insumo.Nombre == req.Nombre {
			return nil, repository.ErrNombreInsumoEnUso
		}
	}
	now := time.Now()
	insumo := models.Insumo{ID: r.nextID, Nombre: req.Nombre, UnidadMedida: req.UnidadMedida,
		StockActual: req.StockActual, StockMinimo: req.StockMinimo, Activo: true, CreatedAt: now, UpdatedAt: now}
	r.insumos[insumo.ID] = insumo
	r.nextID++
	return &insumo, nil
}

func (r *fakeInsumoControllerRepository) GetByID(ctx context.Context, id int) (*models.Insumo, error) {
	insumo, ok := r.insumos[id]
	if !ok {
		return nil, repository.ErrInsumoNoEncontrado
	}
	return &insumo, nil
}

func (r *fakeInsumoControllerRepository) List(ctx context.Context) ([]models.Insumo, error) {
	ids := make([]int, 0, len(r.insumos))
	for id := range r.insumos {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	out := make([]models.Insumo, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.insumos[id])
	}
	return out, nil
}

func (r *fakeInsumoControllerRepository) Update(ctx context.Context, id int, req *models.InsumoUpdateRequest) (*models.Insumo, error) {
	insumo, ok := r.insumos[id]
	if !ok {
		return nil, repository.ErrInsumoNoEncontrado
	}
	if req.Nombre != nil {
		insumo.Nombre = *req.Nombre
	}
	if req.UnidadMedida != nil {
		insumo.UnidadMedida = *req.UnidadMedida
	}
	if req.StockActual != nil {
		insumo.StockActual = *req.StockActual
	}
	if req.StockMinimo != nil {
		insumo.StockMinimo = *req.StockMinimo
	}
	r.insumos[id] = insumo
	return &insumo, nil
}

func (r *fakeInsumoControllerRepository) Delete(ctx context.Context, id int) error {
	insumo, ok := r.insumos[id]
	if !ok {
		return repository.ErrInsumoNoEncontrado
	}
	insumo.Activo = false
	r.insumos[id] = insumo
	return nil
}

func TestInsumoControllerCrearReturnsCreated(t *testing.T) {
	controller := newInsumoControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/insumos", strings.NewReader(`{"nombre":"Detergente","unidad_medida":"litros","stock_actual":10,"stock_minimo":2}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
	var body models.Insumo
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Nombre != "Detergente" || body.StockActual != 10 {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestInsumoControllerCrearReturnsConflictOnDuplicateNombre(t *testing.T) {
	controller := newInsumoControllerForTest(insumoControllerFixture(1, "Detergente"))
	req := httptest.NewRequest(http.MethodPost, "/api/insumos", strings.NewReader(`{"nombre":"Detergente","unidad_medida":"litros"}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestInsumoControllerListarReturnsInsumos(t *testing.T) {
	controller := newInsumoControllerForTest(insumoControllerFixture(1, "Detergente"))
	res := httptest.NewRecorder()

	controller.Listar(res, httptest.NewRequest(http.MethodGet, "/api/insumos", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body []models.Insumo
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 1 || body[0].Nombre != "Detergente" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestInsumoControllerObtenerReturnsNotFound(t *testing.T) {
	controller := newInsumoControllerForTest()
	res := httptest.NewRecorder()

	controller.Obtener(res, withIDParam(http.MethodGet, "/api/insumos/99", "", 99))

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestInsumoControllerActualizarReturnsUpdatedStock(t *testing.T) {
	controller := newInsumoControllerForTest(insumoControllerFixture(1, "Detergente"))
	res := httptest.NewRecorder()

	controller.Actualizar(res, withIDParam(http.MethodPut, "/api/insumos/1", `{"stock_actual":25}`, 1))

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
	}
	var body models.Insumo
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.StockActual != 25 {
		t.Fatalf("expected stock_actual 25, got %+v", body)
	}
}

func TestInsumoControllerEliminarReturnsNoContent(t *testing.T) {
	controller := newInsumoControllerForTest(insumoControllerFixture(1, "Detergente"))
	res := httptest.NewRecorder()

	controller.Eliminar(res, withIDParam(http.MethodDelete, "/api/insumos/1", "", 1))

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}
