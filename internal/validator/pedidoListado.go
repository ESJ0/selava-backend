package validator

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ESJ0/selava-backend/internal/models"
)

// Las fechas de recepción corresponden al horario de Guatemala, UTC-6.
func ParsePedidoList(req models.PedidoListRequest) (models.PedidoListFilter, ValidationErrors) {
	filter := models.PedidoListFilter{Buscar: strings.TrimSpace(req.Buscar), Pagina: 1, Limite: 20, Orden: "recientes"}
	var errs ValidationErrors
	if utf8.RuneCountInString(filter.Buscar) > 100 {
		errs = append(errs, FieldError{Field: "q", Message: "no puede superar los 100 caracteres"})
	}
	parseInt := func(raw, field string, fallback, max int) int {
		if raw == "" {
			return fallback
		}
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > max {
			errs = append(errs, FieldError{Field: field, Message: "debe ser un entero positivo dentro del rango permitido"})
			return fallback
		}
		return value
	}
	filter.EstadoID = parseInt(req.EstadoID, "estado_id", 0, 2147483647)
	filter.Pagina = parseInt(req.Pagina, "pagina", 1, 1000000)
	filter.Limite = parseInt(req.Limite, "limite", 20, 100)
	if req.Orden != "" {
		if req.Orden != "recientes" && req.Orden != "antiguos" {
			errs = append(errs, FieldError{Field: "orden", Message: "debe ser recientes o antiguos"})
		} else {
			filter.Orden = req.Orden
		}
	}
	location := time.FixedZone("America/Guatemala", -6*60*60)
	parseDate := func(raw, field string) *time.Time {
		if raw == "" {
			return nil
		}
		value, err := time.ParseInLocation(time.DateOnly, raw, location)
		if err != nil || value.Year() < 1 {
			errs = append(errs, FieldError{Field: field, Message: "debe ser una fecha válida con formato AAAA-MM-DD"})
			return nil
		}
		return &value
	}
	filter.FechaDesde = parseDate(req.FechaDesde, "fecha_desde")
	hasta := parseDate(req.FechaHasta, "fecha_hasta")
	if hasta != nil {
		if filter.FechaDesde != nil && filter.FechaDesde.After(*hasta) {
			errs = append(errs, FieldError{Field: "fecha_hasta", Message: "no puede ser anterior a fecha_desde"})
		}
		exclusive := hasta.AddDate(0, 0, 1)
		filter.FechaHastaExclusiva = &exclusive
	}
	return filter, errs
}
