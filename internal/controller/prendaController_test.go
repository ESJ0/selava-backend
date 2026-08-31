package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	servicelayer "github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type fakePrendaControllerRepo struct {
	addResult *models.PrendaServicio
	addErr    error
	removeErr error
}

func (r *fakePrendaControllerRepo) CreateMany(context.Context, int, []models.PrendaCreateRequest) ([]models.Prenda, error) {
	return nil, nil
}

func (r *fakePrendaControllerRepo) AddServicio(context.Context, int, int) (*models.PrendaServicio, error) {
	return r.addResult, r.addErr
}

func (r *fakePrendaControllerRepo) RemoveServicio(context.Context, int, int) error {
	return r.removeErr
}

func newPrendaControllerForTest(repo *fakePrendaControllerRepo) *PrendaController {
	return NewPrendaController(servicelayer.NewPrendaService(repo))
}

func servePrendaServicio(controller *PrendaController, method, path, body string) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	router.Post("/api/prendas/{prendaID}/servicios", controller.AsociarServicio)
	router.Delete("/api/prendas/{prendaID}/servicios/{servicioID}", controller.QuitarServicio)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func TestPrendaControllerAsociarServicioReturnsCreated(t *testing.T) {
	esperada := &models.PrendaServicio{ID: 8, PrendaID: 2, ServicioID: 3, PrecioAplicado: 40}
	controller := newPrendaControllerForTest(&fakePrendaControllerRepo{addResult: esperada})

	res := servePrendaServicio(controller, http.MethodPost, "/api/prendas/2/servicios", `{"servicio_id":3}`)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
	var body models.PrendaServicio
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.ID != 8 || body.PrecioAplicado != 40 {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestPrendaControllerAsociarServicioRejectsInvalidServicio(t *testing.T) {
	controller := newPrendaControllerForTest(&fakePrendaControllerRepo{})

	res := servePrendaServicio(controller, http.MethodPost, "/api/prendas/2/servicios", `{"servicio_id":0}`)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, res.Code, res.Body.String())
	}
}

func TestPrendaControllerAsociarServicioReturnsConflict(t *testing.T) {
	controller := newPrendaControllerForTest(&fakePrendaControllerRepo{addErr: repository.ErrPrendaServicioYaAsociado})

	res := servePrendaServicio(controller, http.MethodPost, "/api/prendas/2/servicios", `{"servicio_id":3}`)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, res.Code, res.Body.String())
	}
}

func TestPrendaControllerQuitarServicioReturnsNoContent(t *testing.T) {
	controller := newPrendaControllerForTest(&fakePrendaControllerRepo{})

	res := servePrendaServicio(controller, http.MethodDelete, "/api/prendas/2/servicios/3", "")

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, res.Code, res.Body.String())
	}
}

func TestPrendaControllerQuitarServicioReturnsNotFound(t *testing.T) {
	controller := newPrendaControllerForTest(&fakePrendaControllerRepo{removeErr: repository.ErrPrendaServicioNoEncontrado})

	res := servePrendaServicio(controller, http.MethodDelete, "/api/prendas/2/servicios/3", "")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, res.Code, res.Body.String())
	}
}
