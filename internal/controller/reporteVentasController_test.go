package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/service"
)

type fakeReporteVentasControllerRepository struct {
	reporte *models.ReporteVentas
	err     error
}

type fakeReportePedidosEstadoControllerRepository struct {
	reporte *models.ReportePedidosEstado
	err     error
}

type fakeReporteConsumoInsumosControllerRepository struct {
	reporte *models.ReporteConsumoInsumos
	err     error
}

func (r *fakeReporteConsumoInsumosControllerRepository) GetByPeriodo(_ context.Context, _, _ time.Time) (*models.ReporteConsumoInsumos, error) {
	return r.reporte, r.err
}

func (r *fakeReportePedidosEstadoControllerRepository) Get(_ context.Context) (*models.ReportePedidosEstado, error) {
	return r.reporte, r.err
}

func (r *fakeReporteVentasControllerRepository) GetByPeriodo(_ context.Context, _, _ time.Time) (*models.ReporteVentas, error) {
	return r.reporte, r.err
}

func TestReporteVentasControllerGenerarReturnsOK(t *testing.T) {
	esperado := &models.ReporteVentas{
		FechaInicio:         "2026-09-01",
		FechaFin:            "2026-09-30",
		CantidadVentas:      2,
		CantidadPagos:       3,
		TotalVentas:         21.25,
		VentasPorDia:        []models.ReporteVentasPorDia{},
		VentasPorMetodoPago: []models.ReporteVentasPorMetodoPago{},
	}
	controller := NewReporteVentasController(service.NewReporteVentasService(&fakeReporteVentasControllerRepository{reporte: esperado}))
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/ventas?fecha_inicio=2026-09-01&fecha_fin=2026-09-30", nil)
	res := httptest.NewRecorder()

	controller.Generar(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var body models.ReporteVentas
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.TotalVentas != 21.25 || body.CantidadVentas != 2 || body.CantidadPagos != 3 {
		t.Fatalf("reporte inesperado: %+v", body)
	}
}

func TestReporteVentasControllerValidaParametros(t *testing.T) {
	tests := []struct {
		nombre string
		query  string
	}{
		{nombre: "falta inicio", query: "fecha_fin=2026-09-30"},
		{nombre: "formato invalido", query: "fecha_inicio=01-09-2026&fecha_fin=2026-09-30"},
		{nombre: "periodo invertido", query: "fecha_inicio=2026-10-01&fecha_fin=2026-09-30"},
	}
	controller := NewReporteVentasController(service.NewReporteVentasService(&fakeReporteVentasControllerRepository{}))
	for _, tt := range tests {
		t.Run(tt.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/reportes/ventas?"+tt.query, nil)
			res := httptest.NewRecorder()
			controller.Generar(res, req)
			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestReporteVentasControllerPropagaErrorInterno(t *testing.T) {
	controller := NewReporteVentasController(service.NewReporteVentasService(&fakeReporteVentasControllerRepository{err: context.DeadlineExceeded}))
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/ventas?fecha_inicio=2026-09-01&fecha_fin=2026-09-30", nil)
	res := httptest.NewRecorder()

	controller.Generar(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestReporteControllerPedidosPorEstadoReturnsOK(t *testing.T) {
	esperado := &models.ReportePedidosEstado{
		TotalPedidos: 3,
		PedidosPorEstado: []models.ReportePedidosEstadoDetalle{
			{Estado: models.EstadoPedido{ID: 1, Nombre: "Recibido", Orden: 1}, CantidadPedidos: 2},
			{Estado: models.EstadoPedido{ID: 2, Nombre: "Entregado", Orden: 3}, CantidadPedidos: 1},
		},
	}
	controller := NewReporteController(
		nil,
		service.NewReportePedidosEstadoService(&fakeReportePedidosEstadoControllerRepository{reporte: esperado}),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/pedidos-por-estado", nil)
	res := httptest.NewRecorder()

	controller.PedidosPorEstado(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var body models.ReportePedidosEstado
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.TotalPedidos != 3 || len(body.PedidosPorEstado) != 2 || body.PedidosPorEstado[0].CantidadPedidos != 2 {
		t.Fatalf("reporte inesperado: %+v", body)
	}
}

func TestReporteControllerPedidosPorEstadoPropagaErrorInterno(t *testing.T) {
	controller := NewReporteController(
		nil,
		service.NewReportePedidosEstadoService(&fakeReportePedidosEstadoControllerRepository{err: context.DeadlineExceeded}),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/pedidos-por-estado", nil)
	res := httptest.NewRecorder()

	controller.PedidosPorEstado(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestReporteControllerConsumoInsumosReturnsOK(t *testing.T) {
	esperado := &models.ReporteConsumoInsumos{
		FechaInicio:         "2026-09-01",
		FechaFin:            "2026-09-30",
		CantidadMovimientos: 3,
		TotalConsumido:      8.75,
		ConsumoPorInsumo: []models.ReporteConsumoInsumoDetalle{
			{InsumoID: 1, Nombre: "Detergente", UnidadMedida: "L", CantidadMovimientos: 3, CantidadConsumida: 8.75},
		},
	}
	controller := NewReporteController(
		nil,
		nil,
		service.NewReporteConsumoInsumosService(&fakeReporteConsumoInsumosControllerRepository{reporte: esperado}),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/reportes/consumo-insumos?fecha_inicio=2026-09-01&fecha_fin=2026-09-30", nil)
	res := httptest.NewRecorder()

	controller.ConsumoInsumos(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var body models.ReporteConsumoInsumos
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.TotalConsumido != 8.75 || body.CantidadMovimientos != 3 || len(body.ConsumoPorInsumo) != 1 {
		t.Fatalf("reporte inesperado: %+v", body)
	}
}

func TestReporteControllerConsumoInsumosValidaParametros(t *testing.T) {
	controller := NewReporteController(nil, nil, service.NewReporteConsumoInsumosService(&fakeReporteConsumoInsumosControllerRepository{}))
	for _, query := range []string{
		"fecha_fin=2026-09-30",
		"fecha_inicio=01-09-2026&fecha_fin=2026-09-30",
		"fecha_inicio=2026-10-01&fecha_fin=2026-09-30",
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/reportes/consumo-insumos?"+query, nil)
		res := httptest.NewRecorder()
		controller.ConsumoInsumos(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("query=%s status=%d, body=%s", query, res.Code, res.Body.String())
		}
	}
}
