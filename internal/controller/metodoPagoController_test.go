package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
)

type fakeMetodoPagoRepository struct {
	metodos map[int]models.MetodoPago
	nextID  int
}

func newMetodoPagoControllerForTest(metodos ...models.MetodoPago) *MetodoPagoController {
	repo := &fakeMetodoPagoRepository{metodos: make(map[int]models.MetodoPago), nextID: 1}
	for _, m := range metodos {
		repo.metodos[m.ID] = m
		if m.ID >= repo.nextID {
			repo.nextID = m.ID + 1
		}
	}
	return NewMetodoPagoController(servicelayer.NewMetodoPagoService(repo))
}

func (r *fakeMetodoPagoRepository) Create(ctx context.Context, req *models.MetodoPagoCreateRequest) (*models.MetodoPago, error) {
	for _, m := range r.metodos {
		if m.Nombre == req.Nombre {
			return nil, repository.ErrNombreMetodoPagoEnUso
		}
	}
	metodo := models.MetodoPago{ID: r.nextID, Nombre: req.Nombre, Activo: true}
	r.metodos[metodo.ID] = metodo
	r.nextID++
	return &metodo, nil
}

func (r *fakeMetodoPagoRepository) GetByID(ctx context.Context, id int) (*models.MetodoPago, error) {
	m, ok := r.metodos[id]
	if !ok {
		return nil, repository.ErrMetodoPagoNoEncontrado
	}
	return &m, nil
}

func (r *fakeMetodoPagoRepository) List(ctx context.Context) ([]models.MetodoPago, error) {
	ids := make([]int, 0, len(r.metodos))
	for id := range r.metodos {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	out := make([]models.MetodoPago, 0, len(r.metodos))
	for _, id := range ids {
		out = append(out, r.metodos[id])
	}
	return out, nil
}

func (r *fakeMetodoPagoRepository) Update(ctx context.Context, id int, req *models.MetodoPagoUpdateRequest) (*models.MetodoPago, error) {
	m, ok := r.metodos[id]
	if !ok {
		return nil, repository.ErrMetodoPagoNoEncontrado
	}
	if req.Nombre != nil {
		m.Nombre = *req.Nombre
	}
	r.metodos[id] = m
	return &m, nil
}

func (r *fakeMetodoPagoRepository) Delete(ctx context.Context, id int) error {
	m, ok := r.metodos[id]
	if !ok {
		return repository.ErrMetodoPagoNoEncontrado
	}
	m.Activo = false
	r.metodos[id] = m
	return nil
}

func TestMetodoPagoControllerCrearReturnsCreated(t *testing.T) {
	controller := newMetodoPagoControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/metodos-pago", strings.NewReader(`{"nombre":"Efectivo"}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
	var body models.MetodoPago
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Nombre != "Efectivo" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestMetodoPagoControllerCrearRejectsNombreVacio(t *testing.T) {
	controller := newMetodoPagoControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/metodos-pago", strings.NewReader(`{"nombre":""}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}
}

func TestMetodoPagoControllerCrearReturnsConflictOnDuplicateNombre(t *testing.T) {
	controller := newMetodoPagoControllerForTest(models.MetodoPago{ID: 1, Nombre: "Efectivo", Activo: true})
	req := httptest.NewRequest(http.MethodPost, "/api/metodos-pago", strings.NewReader(`{"nombre":"Efectivo"}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestMetodoPagoControllerListarReturnsMetodos(t *testing.T) {
	controller := newMetodoPagoControllerForTest(models.MetodoPago{ID: 1, Nombre: "Efectivo", Activo: true})
	req := httptest.NewRequest(http.MethodGet, "/api/metodos-pago", nil)
	res := httptest.NewRecorder()

	controller.Listar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body []models.MetodoPago
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 1 || body[0].Nombre != "Efectivo" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestMetodoPagoControllerObtenerReturnsNotFound(t *testing.T) {
	controller := newMetodoPagoControllerForTest()
	req := withIDParam(http.MethodGet, "/api/metodos-pago/99", "", 99)
	res := httptest.NewRecorder()

	controller.Obtener(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestMetodoPagoControllerEliminarReturnsNoContent(t *testing.T) {
	controller := newMetodoPagoControllerForTest(models.MetodoPago{ID: 1, Nombre: "Efectivo", Activo: true})
	req := withIDParam(http.MethodDelete, "/api/metodos-pago/1", "", 1)
	res := httptest.NewRecorder()

	controller.Eliminar(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}
