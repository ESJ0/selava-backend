package repository

import (
	"errors"
	"testing"
)

func TestCalcularStockSalida(t *testing.T) {
	tests := []struct {
		name        string
		stockActual float64
		cantidad    float64
		esperado    float64
		errEsperado error
	}{
		{
			name:        "descuenta una salida con stock disponible",
			stockActual: 10,
			cantidad:    4.5,
			esperado:    5.5,
		},
		{
			name:        "permite consumir exactamente todo el stock",
			stockActual: 10,
			cantidad:    10,
			esperado:    0,
		},
		{
			name:        "rechaza una salida superior al stock",
			stockActual: 10,
			cantidad:    10.01,
			esperado:    10,
			errEsperado: ErrStockInsuficiente,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obtenido, err := calcularStockSalida(tt.stockActual, tt.cantidad)

			if !errors.Is(err, tt.errEsperado) {
				t.Fatalf("expected error %v, got %v", tt.errEsperado, err)
			}
			if obtenido != tt.esperado {
				t.Fatalf("expected stock %.2f, got %.2f", tt.esperado, obtenido)
			}
		})
	}
}
