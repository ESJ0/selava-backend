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

type fakeClienteRepository struct {
	clientes map[int]models.Cliente
	nextID   int
}

func newClienteControllerForTest(clientes ...models.Cliente) *ClienteController {
	repo := &fakeClienteRepository{
		clientes: make(map[int]models.Cliente),
		nextID:   1,
	}
	for _, cliente := range clientes {
		repo.clientes[cliente.ID] = cliente
		if cliente.ID >= repo.nextID {
			repo.nextID = cliente.ID + 1
		}
	}
	return NewClienteController(servicelayer.NewClienteService(repo))
}

func (r *fakeClienteRepository) Create(ctx context.Context, c *models.ClienteCreateRequest) (*models.Cliente, error) {
	now := time.Now()
	cliente := models.Cliente{
		ID:        r.nextID,
		Nombre:    c.Nombre,
		Apellido:  c.Apellido,
		Telefono:  c.Telefono,
		Email:     c.Email,
		Direccion: c.Direccion,
		Activo:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.clientes[cliente.ID] = cliente
	r.nextID++
	return &cliente, nil
}

func (r *fakeClienteRepository) GetByID(ctx context.Context, id int) (*models.Cliente, error) {
	cliente, ok := r.clientes[id]
	if !ok {
		return nil, repository.ErrClienteNoEncontrado
	}
	return &cliente, nil
}

func (r *fakeClienteRepository) List(ctx context.Context) ([]models.Cliente, error) {
	ids := make([]int, 0, len(r.clientes))
	for id := range r.clientes {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	clientes := make([]models.Cliente, 0, len(r.clientes))
	for _, id := range ids {
		clientes = append(clientes, r.clientes[id])
	}
	return clientes, nil
}

func (r *fakeClienteRepository) Update(ctx context.Context, id int, u *models.ClienteUpdateRequest) (*models.Cliente, error) {
	cliente, ok := r.clientes[id]
	if !ok {
		return nil, repository.ErrClienteNoEncontrado
	}
	if u.Nombre != nil {
		cliente.Nombre = *u.Nombre
	}
	if u.Apellido != nil {
		cliente.Apellido = *u.Apellido
	}
	if u.Telefono != nil {
		cliente.Telefono = *u.Telefono
	}
	if u.Email != nil {
		cliente.Email = *u.Email
	}
	if u.Direccion != nil {
		cliente.Direccion = *u.Direccion
	}
	if u.Activo != nil {
		cliente.Activo = *u.Activo
	}
	cliente.UpdatedAt = time.Now()
	r.clientes[id] = cliente
	return &cliente, nil
}

func (r *fakeClienteRepository) Delete(ctx context.Context, id int) error {
	cliente, ok := r.clientes[id]
	if !ok {
		return repository.ErrClienteNoEncontrado
	}
	cliente.Activo = false
	cliente.UpdatedAt = time.Now()
	r.clientes[id] = cliente
	return nil
}

func (r *fakeClienteRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, cliente := range r.clientes {
		if cliente.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func TestClienteControllerCrearReturnsCreated(t *testing.T) {
	controller := newClienteControllerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/clientes", strings.NewReader(`{
		"nombre":"Ana",
		"apellido":"Gomez",
		"telefono":"+502 5555-1234",
		"email":"ana@example.com",
		"direccion":"Zona 10"
	}`))
	res := httptest.NewRecorder()

	controller.Crear(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.Code)
	}
	var body models.ClienteResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.ID != 1 || body.Email != "ana@example.com" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestClienteControllerListarReturnsClientes(t *testing.T) {
	controller := newClienteControllerForTest(clienteControllerFixture(1, "ana@example.com"))
	req := httptest.NewRequest(http.MethodGet, "/api/clientes", nil)
	res := httptest.NewRecorder()

	controller.Listar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body []models.ClienteResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 1 || body[0].Email != "ana@example.com" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestClienteControllerObtenerReturnsNotFound(t *testing.T) {
	controller := newClienteControllerForTest()
	req := requestWithID(http.MethodGet, "/api/clientes/99", nil, 99)
	res := httptest.NewRecorder()

	controller.Obtener(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestClienteControllerActualizarReturnsUpdatedCliente(t *testing.T) {
	controller := newClienteControllerForTest(clienteControllerFixture(1, "ana@example.com"))
	req := requestWithID(http.MethodPut, "/api/clientes/1", strings.NewReader(`{"nombre":"Ana Maria"}`), 1)
	res := httptest.NewRecorder()

	controller.Actualizar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var body models.ClienteResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Nombre != "Ana Maria" {
		t.Fatalf("expected updated name, got %q", body.Nombre)
	}
}

func TestClienteControllerEliminarReturnsNoContent(t *testing.T) {
	controller := newClienteControllerForTest(clienteControllerFixture(1, "ana@example.com"))
	req := requestWithID(http.MethodDelete, "/api/clientes/1", nil, 1)
	res := httptest.NewRecorder()

	controller.Eliminar(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}

func requestWithID(method, target string, body *strings.Reader, id int) *http.Request {
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("")
	} else {
		reader = body
	}
	req := httptest.NewRequest(method, target, reader)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", strconv.Itoa(id))
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func clienteControllerFixture(id int, email string) models.Cliente {
	now := time.Now()
	return models.Cliente{
		ID:        id,
		Nombre:    "Ana",
		Apellido:  "Gomez",
		Telefono:  "+502 5555-1234",
		Email:     email,
		Direccion: "Zona 10",
		Activo:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
