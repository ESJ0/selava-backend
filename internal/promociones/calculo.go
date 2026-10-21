// Package promociones usa centavos y milésimas exactos. float64 solo es el
// formato numérico de la API existente; nunca participa en cálculos monetarios.
package promociones

import (
	"errors"
	"github.com/ESJ0/selava-backend/internal/models"
	"math/big"
	"strconv"
	"time"
)

var ErrInvalida = errors.New("promocion no elegible o medida/importe invalido")

const MaxCentavos int64 = 9999999999

func Escalar(value float64, scale int64) (int64, error) {
	rational, ok := new(big.Rat).SetString(strconv.FormatFloat(value, 'f', -1, 64))
	if !ok || rational.Sign() < 0 {
		return 0, ErrInvalida
	}
	rational.Mul(rational, big.NewRat(scale, 1))
	if !rational.IsInt() || !rational.Num().IsInt64() {
		return 0, ErrInvalida
	}
	return rational.Num().Int64(), nil
}

func Subtotal(centavos, milesimas int64) (int64, error) {
	if centavos < 0 || milesimas < 0 {
		return 0, ErrInvalida
	}
	result := new(big.Int).Mul(big.NewInt(centavos), big.NewInt(milesimas))
	result.Add(result, big.NewInt(500)).Quo(result, big.NewInt(1000))
	if !result.IsInt64() || result.Int64() > MaxCentavos {
		return 0, ErrInvalida
	}
	return result.Int64(), nil
}

func Aplicar(p models.Promocion, mode models.ModalidadCobro, tipoID int, medida int64, precioNormal int64, base int64, ahora time.Time) (*models.PromocionAplicada, error) {
	requerida, err := Escalar(p.MedidaRequerida, 1000)
	if err != nil || requerida <= 0 {
		return nil, ErrInvalida
	}
	precio, err := Escalar(p.PrecioPromocional, 100)
	if err != nil || precio <= 0 || precio > MaxCentavos {
		return nil, ErrInvalida
	}
	if !p.Activo || ahora.Before(p.FechaInicio) || ahora.After(p.FechaFin) || p.TipoPrendaID != tipoID || p.ModalidadCobro != mode || medida < requerida || (mode == models.ModalidadCobroUnidad && (requerida%1000 != 0 || medida%1000 != 0)) {
		return nil, ErrInvalida
	}
	aplicaciones := medida / requerida
	cubierta := aplicaciones * requerida
	sobrante := medida - cubierta
	normal, err := Subtotal(precioNormal, sobrante)
	if err != nil {
		return nil, err
	}
	promo := new(big.Int).Mul(big.NewInt(precio), big.NewInt(aplicaciones))
	total := new(big.Int).Add(new(big.Int).Set(promo), big.NewInt(normal))
	if !total.IsInt64() || total.Int64() > MaxCentavos {
		return nil, ErrInvalida
	}
	return &models.PromocionAplicada{PromocionID: p.ID, Nombre: p.Nombre, Descripcion: p.Descripcion, TipoPrendaID: tipoID, TipoPrendaNombre: p.TipoPrendaNombre, ModalidadCobroAplicada: mode, MedidaRequerida: p.MedidaRequerida, PrecioPromocional: p.PrecioPromocional, MedidaDisponible: float64(medida) / 1000, Aplicaciones: aplicaciones, MedidaCubierta: float64(cubierta) / 1000, Sobrante: float64(sobrante) / 1000, PrecioNormalAplicado: float64(precioNormal) / 100, SubtotalPromocional: float64(promo.Int64()) / 100, SubtotalNormal: float64(normal) / 100, SubtotalBase: float64(base) / 100, SubtotalTotal: float64(total.Int64()) / 100}, nil
}
