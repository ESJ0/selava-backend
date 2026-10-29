package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

const notificacionesSecret = "notificaciones-test-secret"

type notificacionesRouteRepo struct {
	clienteID      int
	notificacionID int
}

func (r *notificacionesRouteRepo) ListByCliente(_ context.Context, clienteID int) (*models.NotificacionesResponse, error) {
	r.clienteID = clienteID
	return &models.NotificacionesResponse{Notificaciones: []models.Notificacion{{ID: 7, ClienteID: clienteID}}, NoLeidas: 1}, nil
}
func (r *notificacionesRouteRepo) Create(context.Context, int, *int, string) (*models.Notificacion, error) {
	return nil, nil
}
func (r *notificacionesRouteRepo) MarkReadByCliente(_ context.Context, notificacionID, clienteID int) (*models.Notificacion, error) {
	r.clienteID, r.notificacionID = clienteID, notificacionID
	return &models.Notificacion{ID: notificacionID, ClienteID: clienteID, Leida: true}, nil
}
func (r *notificacionesRouteRepo) MarkAllReadByCliente(_ context.Context, clienteID int) (int64, error) {
	r.clienteID = clienteID
	return 3, nil
}
func (r *notificacionesRouteRepo) DeleteByCliente(_ context.Context, notificacionID, clienteID int) error {
	r.clienteID, r.notificacionID = clienteID, notificacionID
	return nil
}

func notificacionesRouter(repo service.NotificacionRepository) http.Handler {
	router := chi.NewRouter()
	RegisterNotificacionesCliente(router, controller.NewNotificacionController(service.NewNotificacionService(repo)), notificacionesSecret)
	return router
}

func notificacionClienteRequest(t *testing.T, method, target string, clienteID int) *http.Request {
	t.Helper()
	token, err := auth.GenerateClienteToken(notificacionesSecret, clienteID)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestRutasNotificacionesUsanClienteDelToken(t *testing.T) {
	tests := []struct {
		method, target string
		status         int
	}{
		{http.MethodGet, "/api/cliente/notificaciones/", http.StatusOK},
		{http.MethodPatch, "/api/cliente/notificaciones/7/leida", http.StatusOK},
		{http.MethodPatch, "/api/cliente/notificaciones/leidas", http.StatusOK},
		{http.MethodDelete, "/api/cliente/notificaciones/7", http.StatusNoContent},
	}
	for _, tc := range tests {
		repo := &notificacionesRouteRepo{}
		res := httptest.NewRecorder()
		notificacionesRouter(repo).ServeHTTP(res, notificacionClienteRequest(t, tc.method, tc.target, 42))
		if res.Code != tc.status || repo.clienteID != 42 {
			t.Fatalf("%s %s status=%d clienteID=%d body=%s", tc.method, tc.target, res.Code, repo.clienteID, res.Body.String())
		}
	}
}

func TestRutasNotificacionesRechazanTokenInterno(t *testing.T) {
	token, err := auth.GenerateToken(notificacionesSecret, 7, 1)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/cliente/notificaciones/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	notificacionesRouter(&notificacionesRouteRepo{}).ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized || !strings.Contains(res.Body.String(), "token invalido") {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestRutaNotificacionAjenaNoRevelaExistencia(t *testing.T) {
	repo := &notificacionesRouteRepoAjena{notificacionesRouteRepo: notificacionesRouteRepo{}}
	res := httptest.NewRecorder()
	notificacionesRouter(repo).ServeHTTP(res, notificacionClienteRequest(t, http.MethodDelete, "/api/cliente/notificaciones/99", 42))
	if res.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

type notificacionesRouteRepoAjena struct{ notificacionesRouteRepo }

func (r *notificacionesRouteRepoAjena) DeleteByCliente(_ context.Context, notificacionID, clienteID int) error {
	r.clienteID, r.notificacionID = clienteID, notificacionID
	return repository.ErrNotificacionNoEncontrada
}
