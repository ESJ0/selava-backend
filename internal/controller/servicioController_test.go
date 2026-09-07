package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

// withIDParam simula el enrutador chi inyectando el parametro "id" en el
// contexto de la request, para poder llamar los handlers directo sin
// levantar un router real.
func withIDParam(method, target string, body string, id int) *http.Request {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", strconv.Itoa(id))
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

type fakeServicioRepository struct {
	servicios map[int]models.Servicio
	nextID    int
}

func newServicioControllerForTest(servicios ...models.Servicio) *ServicioController {
	repo := &fakeServicioRepository{servicios: make(map[int]models.Servicio), nextID: 1}
	for _, s := range servicios {
		repo.servicios[s.ID] = s
		if s.ID >= repo.nextID {
			repo.nextID = s.ID + 1
		}
	}
	return NewServicioController(servicelayer.NewServicioService(repo))
}

func (r *fakeServicioRepository) Create(ctx context.Context, req *models.ServicioCreateRequest) (*models.Servicio, error) {
	for _, s := range r.servicios {
		if s.Nombre == req.Nombre {
			return nil, repository.ErrNombreServicioEnUso
		}
	}
	now := time.Now()
	servicio := models.Servicio{
		ID: r.nextID, Nombre: req.Nombre, Descripcion: req.Descripcion,
		PrecioBase: req.PrecioBase, TiempoEstimadoHoras: req.TiempoEstimadoHoras,
		Activo: true, CreatedAt: now, UpdatedAt: now,
	}
	r.servicios[servicio.ID] = servicio
	r.nextID++
	return &servicio, nil
}

func (r *fakeServicioRepository) GetByID(ctx context.Context, id int) (*models.Servicio, error) {
	s, ok := r.servicios[id]
	if !ok {
		return nil, repository.ErrServicioNoEncontrado
	}
	return &s, nil
}

func (r *fakeServicioRepository) List(ctx context.Context) ([]models.Servicio, error) {
	ids := make([]int, 0, len(r.servicios))
	for id := range r.servicios {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	out := make([]models.Servicio, 0, len(r.servicios))
	for _, id := range ids {
		out = append(out, r.servicios[id])
	}
	return out, nil
}

func (r *fakeServicioRepository) Update(ctx context.Context, id int, req *models.ServicioUpdateRequest) (*models.Servicio, error) {
	s, ok := r.servicios[id]
	if !ok {
		return nil, repository.ErrServicioNoEncontrado
	}
	if req.Nombre != nil {
		s.Nombre = *req.Nombre
	}
	if req.PrecioBase != nil {
		s.PrecioBase = *req.PrecioBase
	}
	r.servicios[id] = s
	return &s, nil
}

func (r *fakeServicioRepository) Delete(ctx context.Context, id int) error {
	s, ok := r.servicios[id]
	if !ok {
		return repository.ErrServicioNoEncontrado
	}
	s.Activo = false
	r.servicios[id] = s
	return nil
}

func servicioControllerFixture(id int, nombre string) models.Servicio {
	now := time.Now()
	return models.Servicio{ID: id, Nombre: nombre, PrecioBase: 50, Activo: true, CreatedAt: now, UpdatedAt: now}
}

func TestServicioControllerCrearReturnsCreated(t *testing.T) {
	controller := newServicioControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/servicios", strings.NewReader(`{"nombre":"Lavado","precio_base":50}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
	var body models.Servicio
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Nombre != "Lavado" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestServicioControllerCrearRejectsPrecioInvalido(t *testing.T) {
	controller := newServicioControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/servicios", strings.NewReader(`{"nombre":"Lavado","precio_base":0}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}
}

func TestServicioControllerCrearReturnsConflictOnDuplicateNombre(t *testing.T) {
	controller := newServicioControllerForTest(servicioControllerFixture(1, "Lavado"))
	req := httptest.NewRequest(http.MethodPost, "/api/servicios", strings.NewReader(`{"nombre":"Lavado","precio_base":50}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestServicioControllerListarReturnsServicios(t *testing.T) {
	controller := newServicioControllerForTest(servicioControllerFixture(1, "Lavado"))
	req := httptest.NewRequest(http.MethodGet, "/api/servicios", nil)
	res := httptest.NewRecorder()

	controller.Listar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body []models.Servicio
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 1 || body[0].Nombre != "Lavado" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestServicioControllerObtenerReturnsNotFound(t *testing.T) {
	controller := newServicioControllerForTest()
	req := withIDParam(http.MethodGet, "/api/servicios/99", "", 99)
	res := httptest.NewRecorder()

	controller.Obtener(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestServicioControllerEliminarReturnsNoContent(t *testing.T) {
	controller := newServicioControllerForTest(servicioControllerFixture(1, "Lavado"))
	req := withIDParam(http.MethodDelete, "/api/servicios/1", "", 1)
	res := httptest.NewRecorder()

	controller.Eliminar(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}
