package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

const pedidoEstadoPermissionSecret = "pedido-estado-permission-secret"

type pedidoEstadoPermissionRepo struct {
	service.PedidoRepository
	calls int
}

func (r *pedidoEstadoPermissionRepo) UpdateEstado(_ context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error) {
	r.calls++
	return &models.PedidoEstadoHistorial{
		ID: 1, PedidoID: pedidoID, EstadoID: req.EstadoID, UsuarioID: usuarioID,
	}, nil
}

func pedidoEstadoPermissionRouter(repo *pedidoEstadoPermissionRepo) http.Handler {
	pc := controller.NewPedidoController(service.NewPedidoService(repo))
	return NewRouter(nil, nil, nil, nil, nil, nil, pc, nil, nil, nil, nil, pedidoEstadoPermissionSecret, "http://localhost:5199")
}

func pedidoEstadoPermissionRequest(t *testing.T, role int) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/pedidos/8/estado", strings.NewReader(`{"estado_id":2}`))
	if role == 0 {
		return req
	}
	token, err := auth.GenerateToken(pedidoEstadoPermissionSecret, 7, role)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestActualizarEstadoPedidoPermiteSoloAdministradorYRecepcionista(t *testing.T) {
	tests := []struct {
		name   string
		role   int
		status int
	}{
		{name: "sin autenticar", role: 0, status: http.StatusUnauthorized},
		{name: "administrador", role: middleware.RolAdministrador, status: http.StatusOK},
		{name: "recepcionista", role: middleware.RolRecepcionista, status: http.StatusOK},
		{name: "operario", role: middleware.RolOperario, status: http.StatusForbidden},
		{name: "rol desconocido", role: 99, status: http.StatusForbidden},
	}

	repo := &pedidoEstadoPermissionRepo{}
	router := pedidoEstadoPermissionRouter(repo)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			router.ServeHTTP(res, pedidoEstadoPermissionRequest(t, tc.role))
			if res.Code != tc.status {
				t.Fatalf("status=%d, esperado=%d, body=%s", res.Code, tc.status, res.Body.String())
			}
		})
	}
	if repo.calls != 2 {
		t.Fatalf("el repositorio recibió %d llamadas, esperado 2", repo.calls)
	}
}
