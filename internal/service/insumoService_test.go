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

type fakeInsumoRepo struct {
	insumos     map[int]models.Insumo
	nextID      int
	createCalls int
}

func newFakeInsumoRepo(insumos ...models.Insumo) *fakeInsumoRepo {
	repo := &fakeInsumoRepo{insumos: make(map[int]models.Insumo), nextID: 1}
	for _, insumo := range insumos {
		repo.insumos[insumo.ID] = insumo
		if insumo.ID >= repo.nextID {
			repo.nextID = insumo.ID + 1
		}
	}
	return repo
}

func (r *fakeInsumoRepo) Create(ctx context.Context, req *models.InsumoCreateRequest) (*models.Insumo, error) {
	r.createCalls++
	for _, insumo := range r.insumos {
		if insumo.Nombre == req.Nombre {
			return nil, repository.ErrNombreInsumoEnUso
		}
	}
	now := time.Now()
	insumo := models.Insumo{
		ID: r.nextID, Nombre: req.Nombre, Descripcion: req.Descripcion,
		UnidadMedida: req.UnidadMedida, StockActual: req.StockActual,
		StockMinimo: req.StockMinimo, Activo: true, CreatedAt: now, UpdatedAt: now,
	}
	r.insumos[insumo.ID] = insumo
	r.nextID++
	return &insumo, nil
}

func (r *fakeInsumoRepo) GetByID(ctx context.Context, id int) (*models.Insumo, error) {
	insumo, ok := r.insumos[id]
	if !ok {
		return nil, repository.ErrInsumoNoEncontrado
	}
	return &insumo, nil
}

func (r *fakeInsumoRepo) List(ctx context.Context) ([]models.Insumo, error) {
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

func (r *fakeInsumoRepo) Update(ctx context.Context, id int, req *models.InsumoUpdateRequest) (*models.Insumo, error) {
	insumo, ok := r.insumos[id]
	if !ok {
		return nil, repository.ErrInsumoNoEncontrado
	}
	if req.Nombre != nil {
		for otherID, other := range r.insumos {
			if otherID != id && other.Nombre == *req.Nombre {
				return nil, repository.ErrNombreInsumoEnUso
			}
		}
		insumo.Nombre = *req.Nombre
	}
	if req.Descripcion != nil {
		insumo.Descripcion = req.Descripcion
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
	if req.Activo != nil {
		insumo.Activo = *req.Activo
	}
	insumo.UpdatedAt = time.Now()
	r.insumos[id] = insumo
	return &insumo, nil
}

func (r *fakeInsumoRepo) Delete(ctx context.Context, id int) error {
	insumo, ok := r.insumos[id]
	if !ok {
		return repository.ErrInsumoNoEncontrado
	}
	insumo.Activo = false
	r.insumos[id] = insumo
	return nil
}

func insumoFixture(id int, nombre string) models.Insumo {
	now := time.Now()
	return models.Insumo{
		ID: id, Nombre: nombre, UnidadMedida: "litros", StockActual: 10,
		StockMinimo: 2, Activo: true, CreatedAt: now, UpdatedAt: now,
	}
}

func TestInsumoServiceCrearTrimsAndPersists(t *testing.T) {
	repo := newFakeInsumoRepo()
	service := NewInsumoService(repo)

	insumo, err := service.CrearInsumo(context.Background(), &models.InsumoCreateRequest{
		Nombre: " Detergente ", UnidadMedida: " litros ", StockActual: 10, StockMinimo: 2,
	})
	if err != nil {
		t.Fatalf("CrearInsumo returned error: %v", err)
	}
	if insumo.Nombre != "Detergente" || insumo.UnidadMedida != "litros" || insumo.ID != 1 {
		t.Fatalf("unexpected insumo: %+v", insumo)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestInsumoServiceCrearRejectsInvalidData(t *testing.T) {
	repo := newFakeInsumoRepo()
	service := NewInsumoService(repo)

	_, err := service.CrearInsumo(context.Background(), &models.InsumoCreateRequest{
		Nombre: " ", UnidadMedida: "litros", StockActual: -1,
	})
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestInsumoServiceCrearRejectsNombreDuplicado(t *testing.T) {
	service := NewInsumoService(newFakeInsumoRepo(insumoFixture(1, "Detergente")))

	_, err := service.CrearInsumo(context.Background(), &models.InsumoCreateRequest{
		Nombre: "Detergente", UnidadMedida: "litros",
	})
	if !errors.Is(err, repository.ErrNombreInsumoEnUso) {
		t.Fatalf("expected ErrNombreInsumoEnUso, got %v", err)
	}
}

func TestInsumoServiceCRUDFlow(t *testing.T) {
	repo := newFakeInsumoRepo(insumoFixture(1, "Detergente"), insumoFixture(2, "Suavizante"))
	service := NewInsumoService(repo)

	insumos, err := service.ListarInsumos(context.Background())
	if err != nil || len(insumos) != 2 {
		t.Fatalf("expected 2 insumos, got %d (err=%v)", len(insumos), err)
	}

	obtenido, err := service.ObtenerInsumo(context.Background(), 1)
	if err != nil || obtenido.Nombre != "Detergente" {
		t.Fatalf("unexpected ObtenerInsumo result: %+v (err=%v)", obtenido, err)
	}

	nuevoStock := 25.0
	actualizado, err := service.ActualizarInsumo(context.Background(), 1, &models.InsumoUpdateRequest{StockActual: &nuevoStock})
	if err != nil || actualizado.StockActual != nuevoStock {
		t.Fatalf("unexpected ActualizarInsumo result: %+v (err=%v)", actualizado, err)
	}

	if err := service.EliminarInsumo(context.Background(), 2); err != nil {
		t.Fatalf("EliminarInsumo returned error: %v", err)
	}
	eliminado, _ := service.ObtenerInsumo(context.Background(), 2)
	if eliminado.Activo {
		t.Fatal("expected insumo to be inactive after delete")
	}
}
