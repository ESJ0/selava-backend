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

type fakeTipoPrendaRepository struct {
	tipos  map[int]models.TipoPrenda
	nextID int
}

func newTipoPrendaControllerForTest(tipos ...models.TipoPrenda) *TipoPrendaController {
	repo := &fakeTipoPrendaRepository{tipos: make(map[int]models.TipoPrenda), nextID: 1}
	for _, tp := range tipos {
		repo.tipos[tp.ID] = tp
		if tp.ID >= repo.nextID {
			repo.nextID = tp.ID + 1
		}
	}
	return NewTipoPrendaController(servicelayer.NewTipoPrendaService(repo))
}

func (r *fakeTipoPrendaRepository) Create(ctx context.Context, req *models.TipoPrendaCreateRequest) (*models.TipoPrenda, error) {
	for _, tp := range r.tipos {
		if tp.Nombre == req.Nombre {
			return nil, repository.ErrNombreTipoPrendaEnUso
		}
	}
	now := time.Now()
	tipo := models.TipoPrenda{ID: r.nextID, Nombre: req.Nombre, Descripcion: req.Descripcion, Activo: true, CreatedAt: now, UpdatedAt: now}
	r.tipos[tipo.ID] = tipo
	r.nextID++
	return &tipo, nil
}

func (r *fakeTipoPrendaRepository) GetByID(ctx context.Context, id int) (*models.TipoPrenda, error) {
	tp, ok := r.tipos[id]
	if !ok {
		return nil, repository.ErrTipoPrendaNoEncontrado
	}
	return &tp, nil
}

func (r *fakeTipoPrendaRepository) List(ctx context.Context) ([]models.TipoPrenda, error) {
	ids := make([]int, 0, len(r.tipos))
	for id := range r.tipos {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	out := make([]models.TipoPrenda, 0, len(r.tipos))
	for _, id := range ids {
		out = append(out, r.tipos[id])
	}
	return out, nil
}

func (r *fakeTipoPrendaRepository) Update(ctx context.Context, id int, req *models.TipoPrendaUpdateRequest) (*models.TipoPrenda, error) {
	tp, ok := r.tipos[id]
	if !ok {
		return nil, repository.ErrTipoPrendaNoEncontrado
	}
	if req.Nombre != nil {
		tp.Nombre = *req.Nombre
	}
	r.tipos[id] = tp
	return &tp, nil
}

func (r *fakeTipoPrendaRepository) Delete(ctx context.Context, id int) error {
	tp, ok := r.tipos[id]
	if !ok {
		return repository.ErrTipoPrendaNoEncontrado
	}
	tp.Activo = false
	r.tipos[id] = tp
	return nil
}

func tipoPrendaControllerFixture(id int, nombre string) models.TipoPrenda {
	now := time.Now()
	return models.TipoPrenda{ID: id, Nombre: nombre, Activo: true, CreatedAt: now, UpdatedAt: now}
}

func TestTipoPrendaControllerCrearReturnsCreated(t *testing.T) {
	controller := newTipoPrendaControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/tipos-prenda", strings.NewReader(`{"nombre":"Camisa"}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
	var body models.TipoPrenda
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Nombre != "Camisa" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestTipoPrendaControllerCrearRejectsNombreVacio(t *testing.T) {
	controller := newTipoPrendaControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/tipos-prenda", strings.NewReader(`{"nombre":""}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}
}

func TestTipoPrendaControllerCrearReturnsConflictOnDuplicateNombre(t *testing.T) {
	controller := newTipoPrendaControllerForTest(tipoPrendaControllerFixture(1, "Camisa"))
	req := httptest.NewRequest(http.MethodPost, "/api/tipos-prenda", strings.NewReader(`{"nombre":"Camisa"}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestTipoPrendaControllerListarReturnsTipos(t *testing.T) {
	controller := newTipoPrendaControllerForTest(tipoPrendaControllerFixture(1, "Camisa"))
	req := httptest.NewRequest(http.MethodGet, "/api/tipos-prenda", nil)
	res := httptest.NewRecorder()

	controller.Listar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body []models.TipoPrenda
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 1 || body[0].Nombre != "Camisa" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestTipoPrendaControllerObtenerReturnsNotFound(t *testing.T) {
	controller := newTipoPrendaControllerForTest()
	req := withIDParam(http.MethodGet, "/api/tipos-prenda/99", "", 99)
	res := httptest.NewRecorder()

	controller.Obtener(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestTipoPrendaControllerEliminarReturnsNoContent(t *testing.T) {
	controller := newTipoPrendaControllerForTest(tipoPrendaControllerFixture(1, "Camisa"))
	req := withIDParam(http.MethodDelete, "/api/tipos-prenda/1", "", 1)
	res := httptest.NewRecorder()

	controller.Eliminar(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}
