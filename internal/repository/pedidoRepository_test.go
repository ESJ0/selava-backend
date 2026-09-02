package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestValidarTransicionEstadoPedido(t *testing.T) {
	tests := []struct {
		name    string
		actual  models.EstadoPedido
		destino models.EstadoPedido
		wantErr error
	}{
		{
			name:    "rechaza retroceso",
			actual:  models.EstadoPedido{Nombre: "En lavado", Orden: 2},
			destino: models.EstadoPedido{Nombre: "Recibido", Orden: 1},
			wantErr: ErrPedidoEstadoRetrocedido,
		},
		{
			name:    "rechaza cambio desde entregado",
			actual:  models.EstadoPedido{Nombre: "Entregado", Orden: 4},
			destino: models.EstadoPedido{Nombre: "Entregado", Orden: 4},
			wantErr: ErrPedidoEstadoFinalizado,
		},
		{
			name:    "rechaza cambio desde cancelado",
			actual:  models.EstadoPedido{Nombre: "Cancelado", Orden: 99},
			destino: models.EstadoPedido{Nombre: "Entregado", Orden: 4},
			wantErr: ErrPedidoCancelado,
		},
		{
			name:    "rechaza cancelado por endpoint de estado",
			actual:  models.EstadoPedido{Nombre: "Recibido", Orden: 1},
			destino: models.EstadoPedido{Nombre: "Cancelado", Orden: 99},
			wantErr: ErrCambioEstadoCanceladoNoPermitido,
		},
		{
			name:    "permite avance",
			actual:  models.EstadoPedido{Nombre: "En lavado", Orden: 2},
			destino: models.EstadoPedido{Nombre: "Entregado", Orden: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validarTransicionEstadoPedido(tt.actual, tt.destino); !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidarCancelacionPedido(t *testing.T) {
	tests := []struct {
		nombreEstado string
		wantErr      error
	}{
		{nombreEstado: "Recibido"},
		{nombreEstado: "En lavado", wantErr: ErrPedidoNoCancelable},
		{nombreEstado: "Entregado", wantErr: ErrPedidoNoCancelable},
		{nombreEstado: "Cancelado", wantErr: ErrPedidoNoCancelable},
	}

	for _, tt := range tests {
		t.Run(tt.nombreEstado, func(t *testing.T) {
			if err := validarCancelacionPedido(tt.nombreEstado); !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
