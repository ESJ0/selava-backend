package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
)

type fakeClienteAuthService struct {
	registrado *models.ClienteRegistrado
	token      string
	err        error
}

func (f *fakeClienteAuthService) Registrar(context.Context, *models.ClienteRegistroRequest) (*models.ClienteRegistrado, error) {
	return f.registrado, f.err
}

func (f *fakeClienteAuthService) Login(context.Context, *models.ClienteLoginRequest) (string, error) {
	return f.token, f.err
}

func (f *fakeClienteAuthService) ObtenerPerfil(context.Context, int) (*models.ClienteResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &models.ClienteResponse{ID: 42, Nombre: "Ana", Apellido: "Lopez", Email: "ana@example.com", Activo: true}, nil
}
func (f *fakeClienteAuthService) ActualizarPerfil(context.Context, int, *models.ClientePerfilUpdateRequest) (*models.ClienteResponse, error) {
	return f.ObtenerPerfil(context.Background(), 0)
}
func (f *fakeClienteAuthService) CambiarPassword(context.Context, int, *models.ClientePasswordUpdateRequest) error {
	return f.err
}
func (f *fakeClienteAuthService) GuardarFotoPerfil(context.Context, int, []byte, string) error {
	return f.err
}
func (f *fakeClienteAuthService) ObtenerFotoPerfil(context.Context, int) (*models.ClienteFotoPerfil, error) {
	return &models.ClienteFotoPerfil{Contenido: []byte("foto"), Tipo: "image/png"}, f.err
}
func (f *fakeClienteAuthService) EliminarFotoPerfil(context.Context, int) error { return f.err }

func TestClienteAuthControllerRegistrar(t *testing.T) {
	servicio := &fakeClienteAuthService{registrado: &models.ClienteRegistrado{
		Cliente: models.Cliente{ID: 7, Nombre: "Ana", Activo: true},
		Cuenta:  models.CuentaCliente{ID: 3, ClienteID: 7, Email: "ana@example.com", PasswordHash: "secreto", Activo: true},
	}}
	controller := NewClienteAuthController(servicio)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/clientes/registro", strings.NewReader(`{"nombre":"Ana"}`))
	res := httptest.NewRecorder()

	controller.Registrar(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "password") || !strings.Contains(res.Body.String(), `"cliente_id":7`) {
		t.Fatalf("respuesta insegura o incompleta: %s", res.Body.String())
	}
}

func TestClienteAuthControllerLoginInvalido(t *testing.T) {
	controller := NewClienteAuthController(&fakeClienteAuthService{err: service.ErrCredencialesClienteInvalidas})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/clientes/login", strings.NewReader(`{"email":"ana@example.com","password":"incorrecta"}`))
	res := httptest.NewRecorder()

	controller.Login(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestClienteAuthControllerRegistroDuplicadoRespondeConflicto(t *testing.T) {
	controller := NewClienteAuthController(&fakeClienteAuthService{err: repository.ErrEmailCuentaClienteEnUso})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/clientes/registro", strings.NewReader(`{"nombre":"Ana"}`))
	res := httptest.NewRecorder()

	controller.Registrar(res, req)
	if res.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestClienteAuthControllerRechazaJSONInvalido(t *testing.T) {
	controller := NewClienteAuthController(&fakeClienteAuthService{})
	res := httptest.NewRecorder()
	controller.Registrar(res, httptest.NewRequest(http.MethodPost, "/api/auth/clientes/registro", strings.NewReader(`{`)))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestClienteAuthControllerMeUsaIdentidadDelToken(t *testing.T) {
	const secret = "controller-secret"
	token, err := auth.GenerateClienteToken(secret, 42)
	if err != nil {
		t.Fatal(err)
	}
	controller := NewClienteAuthController(&fakeClienteAuthService{})
	handler := middleware.NewAuthMiddleware(secret).AuthenticateCliente(http.HandlerFunc(controller.Me))
	req := httptest.NewRequest(http.MethodGet, "/api/cliente/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"id":42`) || !strings.Contains(res.Body.String(), `"nombre":"Ana"`) {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestClienteAuthControllerErrorInternoNoFiltraDetalle(t *testing.T) {
	controller := NewClienteAuthController(&fakeClienteAuthService{err: errors.New("detalle sensible")})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/clientes/login", strings.NewReader(`{"email":"ana@example.com","password":"secreto"}`))
	res := httptest.NewRecorder()

	controller.Login(res, req)
	if res.Code != http.StatusInternalServerError || strings.Contains(res.Body.String(), "sensible") {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}
