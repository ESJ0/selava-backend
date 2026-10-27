package repository

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSprint3ActualizarEstadoRegistraHistorialCompleto(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	r := NewPedidoRepository(f.db)
	note := "  inició proceso  "
	// La fecha del historial procede de PostgreSQL; Docker y Windows pueden
	// tener relojes distintos. Comparar contra la misma fuente temporal.
	var start time.Time
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT clock_timestamp()`).Scan(&start))
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
	if len(history) != 2 || history[0].Estado.Nombre != "Recibido" || history[1].Estado.Nombre != "Rackeado" || history[1].Usuario.Nombre != "Usuario" || history[1].Usuario.Apellido != "Prueba" {
		t.Fatalf("historial enriquecido inesperado: %+v", history)
	}
}

func TestSprint3HistorialCronologicoYSaltoAdelanteRechazado(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	r := NewPedidoRepository(f.db)
	_, err := r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Finalizado")}, f.userID)
	if !errors.Is(err, ErrTransicionEstadoNoPermitida) {
		t.Fatalf("salto inesperadamente permitido: %v", err)
	}
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Rackeado")}, f.userID)
	mustSprint3(t, err)
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Finalizado")}, f.userID)
	mustSprint3(t, err)
	history, err := r.GetHistorialEstados(f.ctx, p.ID)
	mustSprint3(t, err)
	if len(history) != 3 || history[0].Estado.Nombre != "Recibido" || history[1].Estado.Nombre != "Rackeado" || history[2].Estado.Nombre != "Finalizado" {
		t.Fatalf("orden inesperado: %+v", history)
	}
	if history[2].FechaCambio.Before(history[1].FechaCambio) || history[1].FechaCambio.Before(history[0].FechaCambio) {
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
	if count != 2 {
		t.Fatalf("rechazos dejaron historial: %d", count)
	}
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Finalizado")}, f.userID)
	mustSprint3(t, err)
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Finalizado")}, f.userID)
	if !errors.Is(err, ErrPedidoEstadoFinalizado) {
		t.Fatalf("finalizado: %v", err)
	}
	cancelled := f.order(t)
	_, err = r.Cancelar(f.ctx, cancelled.ID, f.userID)
	mustSprint3(t, err)
	_, err = r.UpdateEstado(f.ctx, cancelled.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Finalizado")}, f.userID)
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

func TestSEL110FlujoIncluyeSoloEtapasDeServiciosContratados(t *testing.T) {
	f := newSprint3Fixture(t)
	var lavadoID int
	mustSprint3(t, f.db.QueryRow(f.ctx,
		`INSERT INTO servicios(nombre, precio_base) VALUES('Lavado', 1) RETURNING id`,
	).Scan(&lavadoID))

	p, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{
		ClienteID: f.clientID,
		Prendas: []models.PrendaCreateRequest{{
			TipoPrendaID: f.typeID,
			Cantidad:     1,
			Servicios:    []models.PrendaServicioCreateRequest{{ServicioID: lavadoID}},
		}},
	}, f.userID)
	mustSprint3(t, err)

	r := NewPedidoRepository(f.db)
	detail, err := r.GetDetalle(f.ctx, p.ID)
	mustSprint3(t, err)
	nombres := make([]string, 0, len(detail.FlujoOperativo.Estados))
	for _, estado := range detail.FlujoOperativo.Estados {
		nombres = append(nombres, estado.Nombre)
	}
	want := []string{"Recibido", "En lavado", "Rackeado", "Finalizado"}
	if fmt.Sprint(nombres) != fmt.Sprint(want) || detail.FlujoOperativo.SiguienteEstado == nil || detail.FlujoOperativo.SiguienteEstado.Nombre != "En lavado" || detail.FlujoOperativo.ProgresoPorcentaje != 25 {
		t.Fatalf("flujo inesperado: nombres=%v flujo=%+v", nombres, detail.FlujoOperativo)
	}

	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Rackeado")}, f.userID)
	if !errors.Is(err, ErrTransicionEstadoNoPermitida) {
		t.Fatalf("permitió saltar etapa de lavado: %v", err)
	}
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "En lavado")}, f.userID)
	mustSprint3(t, err)
}
