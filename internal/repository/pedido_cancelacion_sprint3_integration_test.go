package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSprint3CancelarRecibidoActualizaEstadoEHistorial(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	r := NewPedidoRepository(f.db)
	cancelled, err := r.Cancelar(f.ctx, p.ID, f.userID)
	mustSprint3(t, err)
	if cancelled.EstadoActualID != f.state(t, "Cancelado") {
		t.Fatalf("estado=%d", cancelled.EstadoActualID)
	}
	history, err := r.GetHistorialEstados(f.ctx, p.ID)
	mustSprint3(t, err)
	if len(history) != 1 || history[0].Estado.Nombre != "Cancelado" || history[0].Usuario.ID != f.userID || history[0].FechaCambio.IsZero() {
		t.Fatalf("historial de cancelación: %+v", history)
	}
}

func TestSprint3CancelarRechazaTodosLosEstadosNoRecibidos(t *testing.T) {
	for _, state := range []string{"Rackeado", "Entregado", "Cancelado"} {
		t.Run(state, func(t *testing.T) {
			f := newSprint3Fixture(t)
			p := f.order(t)
			r := NewPedidoRepository(f.db)
			if state == "Cancelado" {
				_, _ = r.Cancelar(f.ctx, p.ID, f.userID)
			} else {
				_, _ = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, state)}, f.userID)
			}
			_, err := r.Cancelar(f.ctx, p.ID, f.userID)
			if !errors.Is(err, ErrPedidoNoCancelable) {
				t.Fatalf("%s: %v", state, err)
			}
		})
	}
}

func TestSprint3CancelarPedidoYUsuarioInexistentes(t *testing.T) {
	f := newSprint3Fixture(t)
	r := NewPedidoRepository(f.db)
	_, err := r.Cancelar(f.ctx, 2147483647, f.userID)
	if !errors.Is(err, ErrPedidoNoEncontrado) {
		t.Fatalf("pedido inexistente: %v", err)
	}
	p := f.order(t)
	_, err = r.Cancelar(f.ctx, p.ID, 2147483647)
	if !errors.Is(err, ErrUsuarioNoEncontrado) {
		t.Fatalf("usuario inexistente: %v", err)
	}
	var current int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT estado_actual_id FROM pedidos WHERE id=$1`, p.ID).Scan(&current))
	if current != f.state(t, "Recibido") {
		t.Fatal("cancelación fallida alteró el pedido")
	}
}
