package controller

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ESJ0/selava-backend/internal/service"
)

type ReporteController struct {
	ventasService         *service.ReporteVentasService
	pedidosEstadoService  *service.ReportePedidosEstadoService
	consumoInsumosService *service.ReporteConsumoInsumosService
}

// ReporteVentasController se conserva como alias para mantener compatible el
// constructor especifico introducido con el primer reporte.
type ReporteVentasController = ReporteController

func NewReporteVentasController(service *service.ReporteVentasService) *ReporteVentasController {
	return &ReporteController{ventasService: service}
}

func NewReporteController(ventasService *service.ReporteVentasService, pedidosEstadoService *service.ReportePedidosEstadoService, consumoInsumosServices ...*service.ReporteConsumoInsumosService) *ReporteController {
	controller := &ReporteController{
		ventasService:        ventasService,
		pedidosEstadoService: pedidosEstadoService,
	}
	if len(consumoInsumosServices) > 0 {
		controller.consumoInsumosService = consumoInsumosServices[0]
	}
	return controller
}

func (rc *ReporteController) Generar(w http.ResponseWriter, r *http.Request) {
	fechaInicio, ok := fechaReporte(w, r, "fecha_inicio")
	if !ok {
		return
	}
	fechaFin, ok := fechaReporte(w, r, "fecha_fin")
	if !ok {
		return
	}
	if fechaInicio.After(fechaFin) {
		respondError(w, http.StatusBadRequest, "fecha_inicio no puede ser posterior a fecha_fin")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	reporte, err := rc.ventasService.Generar(ctx, fechaInicio, fechaFin)
	if err != nil {
		if errors.Is(err, service.ErrPeriodoReporteInvalido) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	respondJSON(w, http.StatusOK, reporte)
}

func (rc *ReporteController) PedidosPorEstado(w http.ResponseWriter, r *http.Request) {
	if rc.pedidosEstadoService == nil {
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	reporte, err := rc.pedidosEstadoService.Generar(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	respondJSON(w, http.StatusOK, reporte)
}

func (rc *ReporteController) ConsumoInsumos(w http.ResponseWriter, r *http.Request) {
	fechaInicio, ok := fechaReporte(w, r, "fecha_inicio")
	if !ok {
		return
	}
	fechaFin, ok := fechaReporte(w, r, "fecha_fin")
	if !ok {
		return
	}
	if fechaInicio.After(fechaFin) {
		respondError(w, http.StatusBadRequest, "fecha_inicio no puede ser posterior a fecha_fin")
		return
	}
	if rc.consumoInsumosService == nil {
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	reporte, err := rc.consumoInsumosService.Generar(ctx, fechaInicio, fechaFin)
	if err != nil {
		if errors.Is(err, service.ErrPeriodoReporteInvalido) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	respondJSON(w, http.StatusOK, reporte)
}

func fechaReporte(w http.ResponseWriter, r *http.Request, parametro string) (time.Time, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(parametro))
	if value == "" {
		respondError(w, http.StatusBadRequest, parametro+" es requerido")
		return time.Time{}, false
	}
	fecha, err := time.Parse(time.DateOnly, value)
	if err != nil {
		respondError(w, http.StatusBadRequest, parametro+" debe tener formato YYYY-MM-DD")
		return time.Time{}, false
	}
	return fecha, true
}
