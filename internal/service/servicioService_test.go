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

type fakeServicioRepo struct {
	servicios   map[int]models.Servicio
	nextID      int
	createCalls int
}

func newFakeServicioRepo(servicios ...models.Servicio) *fakeServicioRepo {
	repo := &fakeServicioRepo{servicios: make(map[int]models.Servicio), nextID: 1}
	for _, s := range servicios {
		repo.servicios[s.ID] = s
		if s.ID >= repo.nextID {
			repo.nextID = s.ID + 1
		}
	}
	return repo
}

func (r *fakeServicioRepo) Create(ctx context.Context, req *models.ServicioCreateRequest) (*models.Servicio, error) {
	r.createCalls++
	for _, s := range r.servicios {
		if s.Nombre == req.Nombre {
			return nil, repository.ErrNombreServicioEnUso
		}
	}
	now := time.Now()
	servicio := models.Servicio{
		ID:                  r.nextID,
		Nombre:              req.Nombre,
		Descripcion:         req.Descripcion,
		PrecioBase:          req.PrecioBase,
		TiempoEstimadoHoras: req.TiempoEstimadoHoras,
		Activo:              true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	r.servicios[servicio.ID] = servicio
	r.nextID++
	return &servicio, nil
}

func (r *fakeServicioRepo) GetByID(ctx context.Context, id int) (*models.Servicio, error) {
	s, ok := r.servicios[id]
	if !ok {
		return nil, repository.ErrServicioNoEncontrado
	}
	return &s, nil
}

func (r *fakeServicioRepo) List(ctx context.Context) ([]models.Servicio, error) {
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

func (r *fakeServicioRepo) Update(ctx context.Context, id int, req *models.ServicioUpdateRequest) (*models.Servicio, error) {
	s, ok := r.servicios[id]
	if !ok {
		return nil, repository.ErrServicioNoEncontrado
	}
	if req.Nombre != nil {
		for otherID, other := range r.servicios {
			if otherID != id && other.Nombre == *req.Nombre {
				return nil, repository.ErrNombreServicioEnUso
			}
		}
		s.Nombre = *req.Nombre
	}
	if req.Descripcion != nil {
		s.Descripcion = req.Descripcion
	}
	if req.PrecioBase != nil {
		s.PrecioBase = *req.PrecioBase
	}
	if req.TiempoEstimadoHoras != nil {
		s.TiempoEstimadoHoras = req.TiempoEstimadoHoras
	}
	if req.Activo != nil {
		s.Activo = *req.Activo
	}
	s.UpdatedAt = time.Now()
	r.servicios[id] = s
	return &s, nil
}

func (r *fakeServicioRepo) Delete(ctx context.Context, id int) error {
	s, ok := r.servicios[id]
	if !ok {
		return repository.ErrServicioNoEncontrado
	}
	s.Activo = false
	r.servicios[id] = s
	return nil
}

func servicioFixture(id int, nombre string) models.Servicio {
	now := time.Now()
	precio := 50.0
	return models.Servicio{
		ID:         id,
		Nombre:     nombre,
		PrecioBase: precio,
		Activo:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestServicioServiceCrearServicioTrimsAndPersists(t *testing.T) {
	repo := newFakeServicioRepo()
	service := NewServicioService(repo)

	req := &models.ServicioCreateRequest{Nombre: " Lavado en seco ", PrecioBase: 75}
	servicio, err := service.CrearServicio(context.Background(), req)
	if err != nil {
		t.Fatalf("CrearServicio returned error: %v", err)
	}
	if servicio.Nombre != "Lavado en seco" || servicio.ID != 1 {
		t.Fatalf("unexpected servicio: %+v", servicio)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func TestServicioServiceCrearServicioRejectsNombreVacio(t *testing.T) {
	repo := newFakeServicioRepo()
	service := NewServicioService(repo)

	_, err := service.CrearServicio(context.Background(), &models.ServicioCreateRequest{Nombre: "   ", PrecioBase: 10})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestServicioServiceCrearServicioRejectsPrecioInvalido(t *testing.T) {
	repo := newFakeServicioRepo()
	service := NewServicioService(repo)

	_, err := service.CrearServicio(context.Background(), &models.ServicioCreateRequest{Nombre: "Planchado", PrecioBase: 0})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
}

func TestServicioServiceCrearServicioRejectsNombreDuplicado(t *testing.T) {
	repo := newFakeServicioRepo(servicioFixture(1, "Lavado"))
	service := NewServicioService(repo)

	_, err := service.CrearServicio(context.Background(), &models.ServicioCreateRequest{Nombre: "Lavado", PrecioBase: 10})
	if !errors.Is(err, repository.ErrNombreServicioEnUso) {
		t.Fatalf("expected ErrNombreServicioEnUso, got %v", err)
	}
}

func TestServicioServiceCRUDFlow(t *testing.T) {
	repo := newFakeServicioRepo(servicioFixture(1, "Lavado"), servicioFixture(2, "Planchado"))
	service := NewServicioService(repo)

	servicios, err := service.ListarServicios(context.Background())
	if err != nil || len(servicios) != 2 {
		t.Fatalf("expected 2 servicios, got %d (err=%v)", len(servicios), err)
	}

	obtenido, err := service.ObtenerServicio(context.Background(), 1)
	if err != nil || obtenido.Nombre != "Lavado" {
		t.Fatalf("unexpected ObtenerServicio result: %+v (err=%v)", obtenido, err)
	}

	nuevoPrecio := 99.5
	actualizado, err := service.ActualizarServicio(context.Background(), 1, &models.ServicioUpdateRequest{PrecioBase: &nuevoPrecio})
	if err != nil || actualizado.PrecioBase != 99.5 {
		t.Fatalf("unexpected ActualizarServicio result: %+v (err=%v)", actualizado, err)
	}

	if err := service.EliminarServicio(context.Background(), 2); err != nil {
		t.Fatalf("EliminarServicio returned error: %v", err)
	}
	eliminado, _ := service.ObtenerServicio(context.Background(), 2)
	if eliminado.Activo {
		t.Fatal("expected servicio to be inactive after delete")
	}
}

func TestServicioServiceEliminarServicioNotFound(t *testing.T) {
	service := NewServicioService(newFakeServicioRepo())

	err := service.EliminarServicio(context.Background(), 99)
	if !errors.Is(err, repository.ErrServicioNoEncontrado) {
		t.Fatalf("expected ErrServicioNoEncontrado, got %v", err)
	}
}
