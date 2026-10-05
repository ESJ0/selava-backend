package service

import (
	"context"

	"github.com/ESJ0/selava-backend/internal/models"
)

type ReportePedidosEstadoRepository interface {
	Get(ctx context.Context) (*models.ReportePedidosEstado, error)
}

type ReportePedidosEstadoService struct {
	repo ReportePedidosEstadoRepository
}

func NewReportePedidosEstadoService(repo ReportePedidosEstadoRepository) *ReportePedidosEstadoService {
	return &ReportePedidosEstadoService{repo: repo}
}

func (s *ReportePedidosEstadoService) Generar(ctx context.Context) (*models.ReportePedidosEstado, error) {
	return s.repo.Get(ctx)
}
