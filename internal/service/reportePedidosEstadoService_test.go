package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

type fakeReportePedidosEstadoRepository struct {
	reporte *models.ReportePedidosEstado
	err     error
	inicio  time.Time
	fin     time.Time
}

func (r *fakeReportePedidosEstadoRepository) GetByPeriodo(_ context.Context, inicio, fin time.Time) (*models.ReportePedidosEstado, error) {
	r.inicio, r.fin = inicio, fin
	return r.reporte, r.err
}

func TestReportePedidosEstadoServiceGenerar(t *testing.T) {
	esperado := &models.ReportePedidosEstado{TotalPedidos: 3}
	repo := &fakeReportePedidosEstadoRepository{reporte: esperado}
	svc := NewReportePedidosEstadoService(repo)

	reporte, err := svc.Generar(context.Background(), time.Date(2026, 9, 1, 15, 0, 0, 0, time.UTC), time.Date(2026, 9, 30, 8, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Generar() error = %v", err)
	}
	if reporte != esperado {
		t.Fatalf("reporte = %+v, esperado %+v", reporte, esperado)
	}
	if repo.inicio.Hour() != 0 || repo.fin.Format(time.DateOnly) != "2026-10-01" {
		t.Fatalf("periodo normalizado inesperado: %v - %v", repo.inicio, repo.fin)
	}
}

func TestReportePedidosEstadoServicePropagaError(t *testing.T) {
	esperado := errors.New("fallo de consulta")
	svc := NewReportePedidosEstadoService(&fakeReportePedidosEstadoRepository{err: esperado})

	_, err := svc.Generar(context.Background(), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, esperado) {
		t.Fatalf("error = %v, esperado %v", err, esperado)
	}
}

func TestReportePedidosEstadoServiceRechazaPeriodoInvertido(t *testing.T) {
	svc := NewReportePedidosEstadoService(&fakeReportePedidosEstadoRepository{})
	_, err := svc.Generar(context.Background(), time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrPeriodoReporteInvalido) {
		t.Fatalf("error = %v, esperado ErrPeriodoReporteInvalido", err)
	}
}
