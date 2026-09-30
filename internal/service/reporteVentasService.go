package service

import (
	"context"
	"errors"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

var ErrPeriodoReporteInvalido = errors.New("el periodo del reporte es invalido")

type ReporteVentasRepository interface {
	GetByPeriodo(ctx context.Context, inicio, finExclusivo time.Time) (*models.ReporteVentas, error)
}

type ReporteVentasService struct{ repo ReporteVentasRepository }

func NewReporteVentasService(repo ReporteVentasRepository) *ReporteVentasService {
	return &ReporteVentasService{repo: repo}
}

func (s *ReporteVentasService) Generar(ctx context.Context, fechaInicio, fechaFin time.Time) (*models.ReporteVentas, error) {
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

func inicioDelDia(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, fecha.Location())
}
