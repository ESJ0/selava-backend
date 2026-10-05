package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

type fakeReporteVentasRepository struct {
	inicio       time.Time
	finExclusivo time.Time
	reporte      *models.ReporteVentas
	err          error
	llamadas     int
}

func (r *fakeReporteVentasRepository) GetByPeriodo(_ context.Context, inicio, finExclusivo time.Time) (*models.ReporteVentas, error) {
	r.llamadas++
	r.inicio = inicio
	r.finExclusivo = finExclusivo
	return r.reporte, r.err
}

func TestReporteVentasServiceGenerarNormalizaPeriodoInclusivo(t *testing.T) {
	esperado := &models.ReporteVentas{TotalVentas: 125.50}
	repo := &fakeReporteVentasRepository{reporte: esperado}
	svc := NewReporteVentasService(repo)

	reporte, err := svc.Generar(
		context.Background(),
		time.Date(2026, time.September, 1, 14, 30, 0, 0, time.UTC),
		time.Date(2026, time.September, 30, 23, 59, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("Generar() error = %v", err)
	}
	if reporte != esperado {
		t.Fatalf("reporte = %+v, esperado %+v", reporte, esperado)
	}
	if got := repo.inicio.Format(time.RFC3339); got != "2026-09-01T00:00:00Z" {
		t.Fatalf("inicio = %s", got)
	}
	if got := repo.finExclusivo.Format(time.RFC3339); got != "2026-10-01T00:00:00Z" {
		t.Fatalf("fin exclusivo = %s", got)
	}
}

func TestReporteVentasServiceRechazaPeriodoInvertido(t *testing.T) {
	repo := &fakeReporteVentasRepository{}
	svc := NewReporteVentasService(repo)

	_, err := svc.Generar(
		context.Background(),
		time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
	)
	if !errors.Is(err, ErrPeriodoReporteInvalido) {
		t.Fatalf("error = %v, esperado ErrPeriodoReporteInvalido", err)
	}
	if repo.llamadas != 0 {
		t.Fatalf("el repositorio fue llamado %d veces", repo.llamadas)
	}
}
