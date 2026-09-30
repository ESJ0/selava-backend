package service

import (
	"context"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

type ReporteConsumoInsumosRepository interface {
	GetByPeriodo(ctx context.Context, inicio, finExclusivo time.Time) (*models.ReporteConsumoInsumos, error)
}

type ReporteConsumoInsumosService struct {
	repo ReporteConsumoInsumosRepository
}

func NewReporteConsumoInsumosService(repo ReporteConsumoInsumosRepository) *ReporteConsumoInsumosService {
	return &ReporteConsumoInsumosService{repo: repo}
}

func (s *ReporteConsumoInsumosService) Generar(ctx context.Context, fechaInicio, fechaFin time.Time) (*models.ReporteConsumoInsumos, error) {
	if fechaInicio.IsZero() || fechaFin.IsZero() {
		return nil, ErrPeriodoReporteInvalido
	}
	inicio := inicioDelDia(fechaInicio)
	fin := inicioDelDia(fechaFin)
	if inicio.After(fin) {
		return nil, ErrPeriodoReporteInvalido
	}
	return s.repo.GetByPeriodo(ctx, inicio, fin.AddDate(0, 0, 1))
}
