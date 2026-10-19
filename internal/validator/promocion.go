package validator

import (
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/promociones"
	"strings"
)

func ValidatePromocion(req *models.PromocionRequest) ValidationErrors {
	var errs ValidationErrors
	if strings.TrimSpace(req.Nombre) == "" || len(req.Nombre) > 100 {
		errs = append(errs, FieldError{Field: "nombre", Message: "es obligatorio y admite hasta 100 caracteres"})
	}
	if len(req.Descripcion) > 500 {
		errs = append(errs, FieldError{Field: "descripcion", Message: "admite hasta 500 caracteres"})
	}
	if req.TipoPrendaID <= 0 {
		errs = append(errs, FieldError{Field: "tipo_prenda_id", Message: "debe ser positivo"})
	}
	if req.ModalidadCobro != models.ModalidadCobroUnidad && req.ModalidadCobro != models.ModalidadCobroLibra && req.ModalidadCobro != models.ModalidadCobroMetro {
		errs = append(errs, FieldError{Field: "modalidad_cobro", Message: "debe ser UNIDAD, LIBRA o METRO"})
	}
	errs = append(errs, validateMedidaPrenda("medida_requerida", req.MedidaRequerida)...)
	measure, err := promociones.Escalar(req.MedidaRequerida, 1000)
	if err == nil && req.ModalidadCobro == models.ModalidadCobroUnidad && measure%1000 != 0 {
		errs = append(errs, FieldError{Field: "medida_requerida", Message: "debe ser entera para UNIDAD"})
	}
	price, err := promociones.Escalar(req.PrecioPromocional, 100)
	if err != nil || price <= 0 || price > promociones.MaxCentavos {
		errs = append(errs, FieldError{Field: "precio_promocional", Message: "debe ser positivo, hasta 99999999.99, con maximo dos decimales"})
	}
	if req.FechaInicio.IsZero() || req.FechaFin.IsZero() || req.FechaInicio.After(req.FechaFin) {
		errs = append(errs, FieldError{Field: "fecha_fin", Message: "requiere inicio y fin validos; inicio no puede ser posterior al fin"})
	}
	return errs
}

func ValidateCotizacion(req *models.PedidoCreateRequest) ValidationErrors {
	var errs ValidationErrors
	if len(req.Prendas) == 0 {
		return ValidationErrors{{Field: "prendas", Message: "incluye al menos una prenda"}}
	}
	for i, p := range req.Prendas {
		errs = append(errs, validatePrendaCreate(i, p)...)
	}
	return errs
}
