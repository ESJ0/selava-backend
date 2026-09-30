package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/middleware"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

const reportePermissionsSecret = "reporte-permissions-test-secret"

type reporteRouteRepository struct{}

type reportePedidosEstadoRouteRepository struct{}

type reporteConsumoInsumosRouteRepository struct{}

func (reporteRouteRepository) GetByPeriodo(_ context.Context, inicio, fin time.Time) (*models.ReporteVentas, error) {
	return &models.ReporteVentas{
		FechaInicio:         inicio.Format(time.DateOnly),
		FechaFin:            fin.AddDate(0, 0, -1).Format(time.DateOnly),
		VentasPorDia:        []models.ReporteVentasPorDia{},
		VentasPorMetodoPago: []models.ReporteVentasPorMetodoPago{},
	}, nil
}

func (reportePedidosEstadoRouteRepository) Get(_ context.Context) (*models.ReportePedidosEstado, error) {
	return &models.ReportePedidosEstado{PedidosPorEstado: []models.ReportePedidosEstadoDetalle{}}, nil
}

func (reporteConsumoInsumosRouteRepository) GetByPeriodo(_ context.Context, inicio, fin time.Time) (*models.ReporteConsumoInsumos, error) {
	return &models.ReporteConsumoInsumos{
		FechaInicio: inicio.Format(time.DateOnly), FechaFin: fin.AddDate(0, 0, -1).Format(time.DateOnly),
		ConsumoPorInsumo: []models.ReporteConsumoInsumoDetalle{},
	}, nil
}

func reportePermissionsRouter() http.Handler {
	reporteController := controller.NewReporteController(
		service.NewReporteVentasService(reporteRouteRepository{}),
		service.NewReportePedidosEstadoService(reportePedidosEstadoRouteRepository{}),
		service.NewReporteConsumoInsumosService(reporteConsumoInsumosRouteRepository{}),
	)
	return NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, reportePermissionsSecret, "", reporteController)
}

func reporteConsumoInsumosRoleRequest(t *testing.T, role int) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(reportePermissionsSecret, 7, role)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/consumo-insumos?fecha_inicio=2026-09-01&fecha_fin=2026-09-30", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func reportePedidosEstadoRoleRequest(t *testing.T, role int) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(reportePermissionsSecret, 7, role)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/pedidos-por-estado", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func reporteRoleRequest(t *testing.T, role int) *http.Request {
	t.Helper()
	token, err := auth.GenerateToken(reportePermissionsSecret, 7, role)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/ventas?fecha_inicio=2026-09-01&fecha_fin=2026-09-30", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestAdministradorPuedeConsultarReporteVentas(t *testing.T) {
	res := httptest.NewRecorder()
	reportePermissionsRouter().ServeHTTP(res, reporteRoleRequest(t, middleware.RolAdministrador))
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusOK, res.Body.String())
	}
}

func TestRecepcionistaNoPuedeConsultarReporteVentas(t *testing.T) {
	res := httptest.NewRecorder()
	reportePermissionsRouter().ServeHTTP(res, reporteRoleRequest(t, middleware.RolRecepcionista))
	if res.Code != http.StatusForbidden {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusForbidden, res.Body.String())
	}
}

func TestReporteVentasRequiereAutenticacion(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/ventas?fecha_inicio=2026-09-01&fecha_fin=2026-09-30", nil)
	reportePermissionsRouter().ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusUnauthorized, res.Body.String())
	}
}

func TestAdministradorPuedeConsultarReportePedidosPorEstado(t *testing.T) {
	res := httptest.NewRecorder()
	reportePermissionsRouter().ServeHTTP(res, reportePedidosEstadoRoleRequest(t, middleware.RolAdministrador))
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusOK, res.Body.String())
	}
}

func TestRecepcionistaNoPuedeConsultarReportePedidosPorEstado(t *testing.T) {
	res := httptest.NewRecorder()
	reportePermissionsRouter().ServeHTTP(res, reportePedidosEstadoRoleRequest(t, middleware.RolRecepcionista))
	if res.Code != http.StatusForbidden {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusForbidden, res.Body.String())
	}
}

func TestAdministradorPuedeConsultarReporteConsumoInsumos(t *testing.T) {
	res := httptest.NewRecorder()
	reportePermissionsRouter().ServeHTTP(res, reporteConsumoInsumosRoleRequest(t, middleware.RolAdministrador))
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusOK, res.Body.String())
	}
}

func TestRecepcionistaNoPuedeConsultarReporteConsumoInsumos(t *testing.T) {
	res := httptest.NewRecorder()
	reportePermissionsRouter().ServeHTTP(res, reporteConsumoInsumosRoleRequest(t, middleware.RolRecepcionista))
	if res.Code != http.StatusForbidden {
		t.Fatalf("status=%d, esperado=%d: %s", res.Code, http.StatusForbidden, res.Body.String())
	}
}
