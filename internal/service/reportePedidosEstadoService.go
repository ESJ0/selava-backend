package service

import (
	"context"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

type ReportePedidosEstadoRepository interface {
	GetByPeriodo(ctx context.Context, inicio, finExclusivo time.Time) (*models.ReportePedidosEstado, error)
}

type ReportePedidosEstadoService struct {
	repo ReportePedidosEstadoRepository
}

func NewReportePedidosEstadoService(repo ReportePedidosEstadoRepository) *ReportePedidosEstadoService {
	return &ReportePedidosEstadoService{repo: repo}
}

func (s *ReportePedidosEstadoService) Generar(ctx context.Context, fechaInicio, fechaFin time.Time) (*models.ReportePedidosEstado, error) {
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
