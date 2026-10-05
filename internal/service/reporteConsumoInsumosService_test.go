package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

type fakeReporteConsumoInsumosRepository struct {
	inicio       time.Time
	finExclusivo time.Time
	reporte      *models.ReporteConsumoInsumos
	llamadas     int
}

func (r *fakeReporteConsumoInsumosRepository) GetByPeriodo(_ context.Context, inicio, finExclusivo time.Time) (*models.ReporteConsumoInsumos, error) {
	r.llamadas++
	r.inicio = inicio
	r.finExclusivo = finExclusivo
	return r.reporte, nil
}

func TestReporteConsumoInsumosServiceGenerarNormalizaPeriodoInclusivo(t *testing.T) {
	esperado := &models.ReporteConsumoInsumos{TotalConsumido: 8.75}
	repo := &fakeReporteConsumoInsumosRepository{reporte: esperado}
	svc := NewReporteConsumoInsumosService(repo)

	reporte, err := svc.Generar(
		context.Background(),
		time.Date(2026, 9, 1, 15, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC),
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

func TestReporteConsumoInsumosServiceRechazaPeriodoInvertido(t *testing.T) {
	repo := &fakeReporteConsumoInsumosRepository{}
	svc := NewReporteConsumoInsumosService(repo)
	_, err := svc.Generar(
		context.Background(),
		time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
	)
	if !errors.Is(err, ErrPeriodoReporteInvalido) {
		t.Fatalf("error = %v, esperado ErrPeriodoReporteInvalido", err)
	}
	if repo.llamadas != 0 {
		t.Fatalf("el repositorio fue llamado %d veces", repo.llamadas)
	}
}
