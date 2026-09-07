package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSprint3ActualizarEstadoRegistraHistorialCompleto(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	r := NewPedidoRepository(f.db)
	note := "  inició proceso  "
	start := time.Now().Add(-time.Second)
	h, err := r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Rackeado"), Observaciones: &note}, f.userID)
	mustSprint3(t, err)
	if h.PedidoID != p.ID || h.UsuarioID != f.userID || h.EstadoID != f.state(t, "Rackeado") || h.FechaCambio.Before(start) {
		t.Fatalf("historial incompleto: %+v", h)
	}
	var current int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT estado_actual_id FROM pedidos WHERE id=$1`, p.ID).Scan(&current))
	if current != h.EstadoID {
		t.Fatalf("estado actual=%d, historial=%d", current, h.EstadoID)
	}
	history, err := r.GetHistorialEstados(f.ctx, p.ID)
	mustSprint3(t, err)
	if len(history) != 1 || history[0].Estado.Nombre != "Rackeado" || history[0].Usuario.Nombre != "Usuario" || history[0].Usuario.Apellido != "Prueba" {
		t.Fatalf("historial enriquecido inesperado: %+v", history)
	}
}

func TestSprint3HistorialCronologicoYSaltoAdelantePermitido(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	r := NewPedidoRepository(f.db)
	// La regla integrada permite saltar estados hacia adelante; la prueba fija
	// el contrato real en vez de inventar una secuencia más restrictiva.
	_, err := r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Entregado")}, f.userID)
	mustSprint3(t, err)
	var older, newer time.Time
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT fecha_cambio FROM pedido_estados_historial WHERE pedido_id=$1`, p.ID).Scan(&older))
	_, err = f.db.Exec(f.ctx, `UPDATE pedido_estados_historial SET fecha_cambio=$1 WHERE pedido_id=$2`, older.Add(time.Hour), p.ID)
	mustSprint3(t, err)
	// Agregamos una entrada histórica anterior para comprobar orden estable.
	_, err = f.db.Exec(f.ctx, `INSERT INTO pedido_estados_historial(pedido_id,estado_id,usuario_id,fecha_cambio) VALUES($1,$2,$3,$4)`, p.ID, f.state(t, "Recibido"), f.userID, older)
	mustSprint3(t, err)
	history, err := r.GetHistorialEstados(f.ctx, p.ID)
	mustSprint3(t, err)
	if len(history) != 2 || history[0].Estado.Nombre != "Recibido" || history[1].Estado.Nombre != "Entregado" {
		t.Fatalf("orden inesperado: %+v", history)
	}
	newer = history[1].FechaCambio
	if newer.Before(history[0].FechaCambio) {
		t.Fatal("historial no está en orden cronológico")
	}
}

func TestSprint3EstadoRechazaRetrocesoYTerminalesSinEfectosParciales(t *testing.T) {
	f := newSprint3Fixture(t)
	r := NewPedidoRepository(f.db)
	p := f.order(t)
	_, err := r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Rackeado")}, f.userID)
	mustSprint3(t, err)
	for _, tc := range []struct {
		name  string
		state int
		want  error
	}{
		{"retroceso", f.state(t, "Recibido"), ErrPedidoEstadoRetrocedido},
		{"cancelado por endpoint de estado", f.state(t, "Cancelado"), ErrCambioEstadoCanceladoNoPermitido},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, got := r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: tc.state}, f.userID)
			if !errors.Is(got, tc.want) {
				t.Fatalf("esperado %v, obtenido %v", tc.want, got)
			}
		})
	}
	var count int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM pedido_estados_historial WHERE pedido_id=$1`, p.ID).Scan(&count))
	if count != 1 {
		t.Fatalf("rechazos dejaron historial: %d", count)
	}
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Entregado")}, f.userID)
	mustSprint3(t, err)
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Entregado")}, f.userID)
	if !errors.Is(err, ErrPedidoEstadoFinalizado) {
		t.Fatalf("entregado: %v", err)
	}
	cancelled := f.order(t)
	_, err = r.Cancelar(f.ctx, cancelled.ID, f.userID)
	mustSprint3(t, err)
	_, err = r.UpdateEstado(f.ctx, cancelled.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Entregado")}, f.userID)
	if !errors.Is(err, ErrPedidoCancelado) {
		t.Fatalf("cancelado: %v", err)
	}
}

func TestSprint3EstadoRechazaPedidoEstadoYUsuarioInexistentes(t *testing.T) {
	f := newSprint3Fixture(t)
	r := NewPedidoRepository(f.db)
	p := f.order(t)
	for _, tc := range []struct {
		name                    string
		pedido, estado, usuario int
		want                    error
	}{
		{"pedido", 2147483647, f.state(t, "Rackeado"), f.userID, ErrPedidoNoEncontrado},
		{"estado", p.ID, 2147483647, f.userID, ErrEstadoPedidoDestinoNoEncontrado},
		{"usuario", p.ID, f.state(t, "Rackeado"), 2147483647, ErrUsuarioNoEncontrado},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := r.UpdateEstado(f.ctx, tc.pedido, &models.PedidoEstadoUpdateRequest{EstadoID: tc.estado}, tc.usuario)
			if !errors.Is(err, tc.want) {
				t.Fatalf("esperado %v, obtenido %v", tc.want, err)
			}
		})
	}
}
