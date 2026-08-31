package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type fakePrendaRepo struct {
	addResult    *models.PrendaServicio
	addErr       error
	removeErr    error
	addCalls     int
	removeCalls  int
	lastPrendaID int
	lastServicio int
}

func (r *fakePrendaRepo) CreateMany(context.Context, int, []models.PrendaCreateRequest) ([]models.Prenda, error) {
	return nil, nil
}

func (r *fakePrendaRepo) AddServicio(_ context.Context, prendaID, servicioID int) (*models.PrendaServicio, error) {
	r.addCalls++
	r.lastPrendaID = prendaID
	r.lastServicio = servicioID
	return r.addResult, r.addErr
}

func (r *fakePrendaRepo) RemoveServicio(_ context.Context, prendaID, servicioID int) error {
	r.removeCalls++
	r.lastPrendaID = prendaID
	r.lastServicio = servicioID
	return r.removeErr
}

func TestPrendaServiceAsociarServicio(t *testing.T) {
	esperada := &models.PrendaServicio{ID: 3, PrendaID: 4, ServicioID: 5, PrecioAplicado: 25}
	repo := &fakePrendaRepo{addResult: esperada}
	servicio := NewPrendaService(repo)

	obtenida, err := servicio.AsociarServicio(context.Background(), 4, &models.PrendaServicioCreateRequest{ServicioID: 5})
	if err != nil {
		t.Fatalf("AsociarServicio returned error: %v", err)
	}
	if obtenida != esperada || repo.addCalls != 1 || repo.lastPrendaID != 4 || repo.lastServicio != 5 {
		t.Fatalf("unexpected association result: %+v, repo: %+v", obtenida, repo)
	}
}

func TestPrendaServiceAsociarServicioRechazaServicioInvalido(t *testing.T) {
	repo := &fakePrendaRepo{}
	servicio := NewPrendaService(repo)

	_, err := servicio.AsociarServicio(context.Background(), 4, &models.PrendaServicioCreateRequest{})

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.addCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.addCalls)
	}
}

func TestPrendaServiceQuitarServicio(t *testing.T) {
	repo := &fakePrendaRepo{}
	servicio := NewPrendaService(repo)

	if err := servicio.QuitarServicio(context.Background(), 7, 9); err != nil {
		t.Fatalf("QuitarServicio returned error: %v", err)
	}
	if repo.removeCalls != 1 || repo.lastPrendaID != 7 || repo.lastServicio != 9 {
		t.Fatalf("unexpected repository call: %+v", repo)
	}
}
