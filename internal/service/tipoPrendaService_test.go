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

type fakeTipoPrendaRepo struct {
	tipos       map[int]models.TipoPrenda
	nextID      int
	createCalls int
}

func newFakeTipoPrendaRepo(tipos ...models.TipoPrenda) *fakeTipoPrendaRepo {
	repo := &fakeTipoPrendaRepo{tipos: make(map[int]models.TipoPrenda), nextID: 1}
	for _, tp := range tipos {
		repo.tipos[tp.ID] = tp
		if tp.ID >= repo.nextID {
			repo.nextID = tp.ID + 1
		}
	}
	return repo
}

func (r *fakeTipoPrendaRepo) Create(ctx context.Context, req *models.TipoPrendaCreateRequest) (*models.TipoPrenda, error) {
	r.createCalls++
	for _, tp := range r.tipos {
		if tp.Nombre == req.Nombre {
			return nil, repository.ErrNombreTipoPrendaEnUso
		}
	}
	now := time.Now()
	tipo := models.TipoPrenda{
		ID:          r.nextID,
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
		Activo:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.tipos[tipo.ID] = tipo
	r.nextID++
	return &tipo, nil
}

func (r *fakeTipoPrendaRepo) GetByID(ctx context.Context, id int) (*models.TipoPrenda, error) {
	tp, ok := r.tipos[id]
	if !ok {
		return nil, repository.ErrTipoPrendaNoEncontrado
	}
	return &tp, nil
}

func (r *fakeTipoPrendaRepo) List(ctx context.Context) ([]models.TipoPrenda, error) {
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

func (r *fakeTipoPrendaRepo) Update(ctx context.Context, id int, req *models.TipoPrendaUpdateRequest) (*models.TipoPrenda, error) {
	tp, ok := r.tipos[id]
	if !ok {
		return nil, repository.ErrTipoPrendaNoEncontrado
	}
	if req.Nombre != nil {
		for otherID, other := range r.tipos {
			if otherID != id && other.Nombre == *req.Nombre {
				return nil, repository.ErrNombreTipoPrendaEnUso
			}
		}
		tp.Nombre = *req.Nombre
	}
	if req.Descripcion != nil {
		tp.Descripcion = req.Descripcion
	}
	if req.Activo != nil {
		tp.Activo = *req.Activo
	}
	tp.UpdatedAt = time.Now()
	r.tipos[id] = tp
	return &tp, nil
}

func (r *fakeTipoPrendaRepo) Delete(ctx context.Context, id int) error {
	tp, ok := r.tipos[id]
	if !ok {
		return repository.ErrTipoPrendaNoEncontrado
	}
	tp.Activo = false
	r.tipos[id] = tp
	return nil
}

func tipoPrendaFixture(id int, nombre string) models.TipoPrenda {
	now := time.Now()
	return models.TipoPrenda{ID: id, Nombre: nombre, Activo: true, CreatedAt: now, UpdatedAt: now}
}

func TestTipoPrendaServiceCrearTrimsAndPersists(t *testing.T) {
	repo := newFakeTipoPrendaRepo()
	service := NewTipoPrendaService(repo)

	tipo, err := service.CrearTipoPrenda(context.Background(), &models.TipoPrendaCreateRequest{Nombre: " Camisa "})
	if err != nil {
		t.Fatalf("CrearTipoPrenda returned error: %v", err)
	}
	if tipo.Nombre != "Camisa" || tipo.ID != 1 {
		t.Fatalf("unexpected tipo: %+v", tipo)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestTipoPrendaServiceCrearRejectsNombreVacio(t *testing.T) {
	repo := newFakeTipoPrendaRepo()
	service := NewTipoPrendaService(repo)

	_, err := service.CrearTipoPrenda(context.Background(), &models.TipoPrendaCreateRequest{Nombre: "  "})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestTipoPrendaServiceCrearRejectsDescripcionLarga(t *testing.T) {
	repo := newFakeTipoPrendaRepo()
	service := NewTipoPrendaService(repo)

	larga := make([]byte, 256)
	for i := range larga {
		larga[i] = 'a'
	}
	descripcion := string(larga)

	_, err := service.CrearTipoPrenda(context.Background(), &models.TipoPrendaCreateRequest{
		Nombre:      "Pantalon",
		Descripcion: &descripcion,
	})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
}

func TestTipoPrendaServiceCrearRejectsNombreDuplicado(t *testing.T) {
	repo := newFakeTipoPrendaRepo(tipoPrendaFixture(1, "Camisa"))
	service := NewTipoPrendaService(repo)

	_, err := service.CrearTipoPrenda(context.Background(), &models.TipoPrendaCreateRequest{Nombre: "Camisa"})
	if !errors.Is(err, repository.ErrNombreTipoPrendaEnUso) {
		t.Fatalf("expected ErrNombreTipoPrendaEnUso, got %v", err)
	}
}

func TestTipoPrendaServiceCRUDFlow(t *testing.T) {
	repo := newFakeTipoPrendaRepo(tipoPrendaFixture(1, "Camisa"), tipoPrendaFixture(2, "Pantalon"))
	service := NewTipoPrendaService(repo)

	tipos, err := service.ListarTiposPrenda(context.Background())
	if err != nil || len(tipos) != 2 {
		t.Fatalf("expected 2 tipos, got %d (err=%v)", len(tipos), err)
	}

	obtenido, err := service.ObtenerTipoPrenda(context.Background(), 1)
	if err != nil || obtenido.Nombre != "Camisa" {
		t.Fatalf("unexpected ObtenerTipoPrenda result: %+v (err=%v)", obtenido, err)
	}

	nuevoNombre := "Camisa manga larga"
	actualizado, err := service.ActualizarTipoPrenda(context.Background(), 1, &models.TipoPrendaUpdateRequest{Nombre: &nuevoNombre})
	if err != nil || actualizado.Nombre != nuevoNombre {
		t.Fatalf("unexpected ActualizarTipoPrenda result: %+v (err=%v)", actualizado, err)
	}

	if err := service.EliminarTipoPrenda(context.Background(), 2); err != nil {
		t.Fatalf("EliminarTipoPrenda returned error: %v", err)
	}
	eliminado, _ := service.ObtenerTipoPrenda(context.Background(), 2)
	if eliminado.Activo {
		t.Fatal("expected tipo de prenda to be inactive after delete")
	}
}

func TestTipoPrendaServiceEliminarNotFound(t *testing.T) {
	service := NewTipoPrendaService(newFakeTipoPrendaRepo())

	err := service.EliminarTipoPrenda(context.Background(), 99)
	if !errors.Is(err, repository.ErrTipoPrendaNoEncontrado) {
		t.Fatalf("expected ErrTipoPrendaNoEncontrado, got %v", err)
	}
}
