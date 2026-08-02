package service

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type fakeClienteRepo struct {
	clientes    map[int]models.Cliente
	nextID      int
	createCalls int
}

func newFakeClienteRepo(clientes ...models.Cliente) *fakeClienteRepo {
	repo := &fakeClienteRepo{
		clientes: make(map[int]models.Cliente),
		nextID:   1,
	}
	for _, cliente := range clientes {
		repo.clientes[cliente.ID] = cliente
		if cliente.ID >= repo.nextID {
			repo.nextID = cliente.ID + 1
		}
	}
	return repo
}

func (r *fakeClienteRepo) Create(ctx context.Context, c *models.ClienteCreateRequest) (*models.Cliente, error) {
	r.createCalls++
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

func (r *fakeClienteRepo) GetByID(ctx context.Context, id int) (*models.Cliente, error) {
	cliente, ok := r.clientes[id]
	if !ok {
		return nil, repository.ErrClienteNoEncontrado
	}
	return &cliente, nil
}

func (r *fakeClienteRepo) List(ctx context.Context) ([]models.Cliente, error) {
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

func (r *fakeClienteRepo) Update(ctx context.Context, id int, u *models.ClienteUpdateRequest) (*models.Cliente, error) {
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

func (r *fakeClienteRepo) Delete(ctx context.Context, id int) error {
	cliente, ok := r.clientes[id]
	if !ok {
		return repository.ErrClienteNoEncontrado
	}
	cliente.Activo = false
	cliente.UpdatedAt = time.Now()
	r.clientes[id] = cliente
	return nil
}

func (r *fakeClienteRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, cliente := range r.clientes {
		if cliente.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func TestClienteServiceCrearClienteTrimsAndPersists(t *testing.T) {
	repo := newFakeClienteRepo()
	service := NewClienteService(repo)
	req := &models.ClienteCreateRequest{
		Nombre:    " Ana ",
		Apellido:  " Gomez ",
		Telefono:  "+502 5555-1234",
		Email:     " ana@example.com ",
		Direccion: "Zona 10",
	}

	cliente, err := service.CrearCliente(context.Background(), req)
	if err != nil {
		t.Fatalf("CrearCliente returned error: %v", err)
	}

	if cliente.ID != 1 || cliente.Nombre != "Ana" || cliente.Apellido != "Gomez" || cliente.Email != "ana@example.com" {
		t.Fatalf("unexpected cliente: %+v", cliente)
	}
	if req.Nombre != "Ana" || req.Apellido != "Gomez" || req.Email != "ana@example.com" {
		t.Fatalf("expected request fields to be trimmed, got %+v", req)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestClienteServiceCrearClienteRejectsDuplicateEmail(t *testing.T) {
	repo := newFakeClienteRepo(clienteFixture(1, "ana@example.com"))
	service := NewClienteService(repo)

	_, err := service.CrearCliente(context.Background(), &models.ClienteCreateRequest{
		Nombre:   "Ana",
		Apellido: "Gomez",
		Telefono: "+502 5555-1234",
		Email:    "ana@example.com",
	})
	if !errors.Is(err, ErrEmailEnUso) {
		t.Fatalf("expected ErrEmailEnUso, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestClienteServiceCrearClienteRejectsInvalidData(t *testing.T) {
	repo := newFakeClienteRepo()
	service := NewClienteService(repo)

	_, err := service.CrearCliente(context.Background(), &models.ClienteCreateRequest{
		Nombre:   "",
		Apellido: "",
		Telefono: "123",
		Email:    "email-invalido",
	})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestClienteServiceCRUDFlow(t *testing.T) {
	repo := newFakeClienteRepo(
		clienteFixture(1, "ana@example.com"),
		clienteFixture(2, "luis@example.com"),
	)
	service := NewClienteService(repo)

	clientes, err := service.ListarClientes(context.Background())
	if err != nil {
		t.Fatalf("ListarClientes returned error: %v", err)
	}
	if len(clientes) != 2 {
		t.Fatalf("expected 2 clientes, got %d", len(clientes))
	}

	obtenido, err := service.ObtenerCliente(context.Background(), 1)
	if err != nil {
		t.Fatalf("ObtenerCliente returned error: %v", err)
	}
	if obtenido.Email != "ana@example.com" {
		t.Fatalf("expected ana@example.com, got %q", obtenido.Email)
	}

	nombre := "Ana Maria"
	actualizado, err := service.ActualizarCliente(context.Background(), 1, &models.ClienteUpdateRequest{
		Nombre: &nombre,
	})
	if err != nil {
		t.Fatalf("ActualizarCliente returned error: %v", err)
	}
	if actualizado.Nombre != "Ana Maria" {
		t.Fatalf("expected updated name, got %q", actualizado.Nombre)
	}

	if err := service.EliminarCliente(context.Background(), 2); err != nil {
		t.Fatalf("EliminarCliente returned error: %v", err)
	}
	eliminado, err := service.ObtenerCliente(context.Background(), 2)
	if err != nil {
		t.Fatalf("ObtenerCliente after delete returned error: %v", err)
	}
	if eliminado.Activo {
		t.Fatal("expected cliente to be inactive after delete")
	}
}

func TestClienteServiceActualizarClienteRejectsEmailUsedByAnotherClient(t *testing.T) {
	repo := newFakeClienteRepo(
		clienteFixture(1, "ana@example.com"),
		clienteFixture(2, "luis@example.com"),
	)
	service := NewClienteService(repo)
	email := "luis@example.com"

	_, err := service.ActualizarCliente(context.Background(), 1, &models.ClienteUpdateRequest{
		Email: &email,
	})
	if !errors.Is(err, ErrEmailEnUso) {
		t.Fatalf("expected ErrEmailEnUso, got %v", err)
	}
}

func TestClienteServiceActualizarClienteAllowsSameEmail(t *testing.T) {
	repo := newFakeClienteRepo(clienteFixture(1, "ana@example.com"))
	service := NewClienteService(repo)
	email := "ana@example.com"

	cliente, err := service.ActualizarCliente(context.Background(), 1, &models.ClienteUpdateRequest{
		Email: &email,
	})
	if err != nil {
		t.Fatalf("ActualizarCliente returned error: %v", err)
	}
	if cliente.Email != "ana@example.com" {
		t.Fatalf("expected same email, got %q", cliente.Email)
	}
}

func TestClienteServiceEliminarClienteNotFound(t *testing.T) {
	service := NewClienteService(newFakeClienteRepo())

	err := service.EliminarCliente(context.Background(), 99)
	if !errors.Is(err, repository.ErrClienteNoEncontrado) {
		t.Fatalf("expected ErrClienteNoEncontrado, got %v", err)
	}
}

func clienteFixture(id int, email string) models.Cliente {
	now := time.Now()
	return models.Cliente{
		ID:        id,
		Nombre:    "Nombre",
		Apellido:  "Apellido",
		Telefono:  "+502 5555-1234",
		Email:     email,
		Direccion: "Zona 10",
		Activo:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
