package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type fakeMetodoPagoRepo struct {
	metodos     map[int]models.MetodoPago
	nextID      int
	createCalls int
}

func newFakeMetodoPagoRepo(metodos ...models.MetodoPago) *fakeMetodoPagoRepo {
	repo := &fakeMetodoPagoRepo{metodos: make(map[int]models.MetodoPago), nextID: 1}
	for _, m := range metodos {
		repo.metodos[m.ID] = m
		if m.ID >= repo.nextID {
			repo.nextID = m.ID + 1
		}
	}
	return repo
}

func (r *fakeMetodoPagoRepo) Create(ctx context.Context, req *models.MetodoPagoCreateRequest) (*models.MetodoPago, error) {
	r.createCalls++
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

func (r *fakeMetodoPagoRepo) GetByID(ctx context.Context, id int) (*models.MetodoPago, error) {
	m, ok := r.metodos[id]
	if !ok {
		return nil, repository.ErrMetodoPagoNoEncontrado
	}
	return &m, nil
}

func (r *fakeMetodoPagoRepo) List(ctx context.Context) ([]models.MetodoPago, error) {
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

func (r *fakeMetodoPagoRepo) Update(ctx context.Context, id int, req *models.MetodoPagoUpdateRequest) (*models.MetodoPago, error) {
	m, ok := r.metodos[id]
	if !ok {
		return nil, repository.ErrMetodoPagoNoEncontrado
	}
	if req.Nombre != nil {
		for otherID, other := range r.metodos {
			if otherID != id && other.Nombre == *req.Nombre {
				return nil, repository.ErrNombreMetodoPagoEnUso
			}
		}
		m.Nombre = *req.Nombre
	}
	if req.Activo != nil {
		m.Activo = *req.Activo
	}
	r.metodos[id] = m
	return &m, nil
}

func (r *fakeMetodoPagoRepo) Delete(ctx context.Context, id int) error {
	m, ok := r.metodos[id]
	if !ok {
		return repository.ErrMetodoPagoNoEncontrado
	}
	m.Activo = false
	r.metodos[id] = m
	return nil
}

func TestMetodoPagoServiceCrearTrimsAndPersists(t *testing.T) {
	repo := newFakeMetodoPagoRepo()
	service := NewMetodoPagoService(repo)

	metodo, err := service.CrearMetodoPago(context.Background(), &models.MetodoPagoCreateRequest{Nombre: " Efectivo "})
	if err != nil {
		t.Fatalf("CrearMetodoPago returned error: %v", err)
	}
	if metodo.Nombre != "Efectivo" || metodo.ID != 1 {
		t.Fatalf("unexpected metodo: %+v", metodo)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestMetodoPagoServiceCrearRejectsNombreVacio(t *testing.T) {
	repo := newFakeMetodoPagoRepo()
	service := NewMetodoPagoService(repo)

	_, err := service.CrearMetodoPago(context.Background(), &models.MetodoPagoCreateRequest{Nombre: "   "})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestMetodoPagoServiceCrearRejectsNombreLargo(t *testing.T) {
	repo := newFakeMetodoPagoRepo()
	service := NewMetodoPagoService(repo)

	nombreLargo := strings.Repeat("a", 51)
	_, err := service.CrearMetodoPago(context.Background(), &models.MetodoPagoCreateRequest{Nombre: nombreLargo})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
}

func TestMetodoPagoServiceCrearRejectsNombreDuplicado(t *testing.T) {
	repo := newFakeMetodoPagoRepo(models.MetodoPago{ID: 1, Nombre: "Efectivo", Activo: true})
	service := NewMetodoPagoService(repo)

	_, err := service.CrearMetodoPago(context.Background(), &models.MetodoPagoCreateRequest{Nombre: "Efectivo"})
	if !errors.Is(err, repository.ErrNombreMetodoPagoEnUso) {
		t.Fatalf("expected ErrNombreMetodoPagoEnUso, got %v", err)
	}
}

func TestMetodoPagoServiceCRUDFlow(t *testing.T) {
	repo := newFakeMetodoPagoRepo(
		models.MetodoPago{ID: 1, Nombre: "Efectivo", Activo: true},
		models.MetodoPago{ID: 2, Nombre: "Tarjeta", Activo: true},
	)
	service := NewMetodoPagoService(repo)

	metodos, err := service.ListarMetodosPago(context.Background())
	if err != nil || len(metodos) != 2 {
		t.Fatalf("expected 2 metodos, got %d (err=%v)", len(metodos), err)
	}

	obtenido, err := service.ObtenerMetodoPago(context.Background(), 1)
	if err != nil || obtenido.Nombre != "Efectivo" {
		t.Fatalf("unexpected ObtenerMetodoPago result: %+v (err=%v)", obtenido, err)
	}

	nuevoNombre := "Efectivo (caja chica)"
	actualizado, err := service.ActualizarMetodoPago(context.Background(), 1, &models.MetodoPagoUpdateRequest{Nombre: &nuevoNombre})
	if err != nil || actualizado.Nombre != nuevoNombre {
		t.Fatalf("unexpected ActualizarMetodoPago result: %+v (err=%v)", actualizado, err)
	}

	if err := service.EliminarMetodoPago(context.Background(), 2); err != nil {
		t.Fatalf("EliminarMetodoPago returned error: %v", err)
	}
	eliminado, _ := service.ObtenerMetodoPago(context.Background(), 2)
	if eliminado.Activo {
		t.Fatal("expected metodo de pago to be inactive after delete")
	}
}

func TestMetodoPagoServiceEliminarNotFound(t *testing.T) {
	service := NewMetodoPagoService(newFakeMetodoPagoRepo())

	err := service.EliminarMetodoPago(context.Background(), 99)
	if !errors.Is(err, repository.ErrMetodoPagoNoEncontrado) {
		t.Fatalf("expected ErrMetodoPagoNoEncontrado, got %v", err)
	}
}
