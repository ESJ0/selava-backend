package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

type fakeReportePedidosEstadoRepository struct {
	reporte *models.ReportePedidosEstado
	err     error
}

func (r *fakeReportePedidosEstadoRepository) Get(_ context.Context) (*models.ReportePedidosEstado, error) {
	return r.reporte, r.err
}

func TestReportePedidosEstadoServiceGenerar(t *testing.T) {
	esperado := &models.ReportePedidosEstado{TotalPedidos: 3}
	svc := NewReportePedidosEstadoService(&fakeReportePedidosEstadoRepository{reporte: esperado})

	reporte, err := svc.Generar(context.Background())
	if err != nil {
		t.Fatalf("Generar() error = %v", err)
	}
	if reporte != esperado {
		t.Fatalf("reporte = %+v, esperado %+v", reporte, esperado)
	}
}

func TestReportePedidosEstadoServicePropagaError(t *testing.T) {
	esperado := errors.New("fallo de consulta")
	svc := NewReportePedidosEstadoService(&fakeReportePedidosEstadoRepository{err: esperado})

	_, err := svc.Generar(context.Background())
	if !errors.Is(err, esperado) {
		t.Fatalf("error = %v, esperado %v", err, esperado)
	}
}
