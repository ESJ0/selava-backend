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
		{"repeticion", f.state(t, "Rackeado"), ErrTransicionEstadoNoPermitida},
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

func TestSEL110HistorialRegistraCadaCambioConFechaYUsuarioResponsable(t *testing.T) {
	f := newSprint3Fixture(t)
	var segundoUsuarioID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `
		INSERT INTO usuarios(rol_id,nombre,apellido,email,password_hash,telefono)
		VALUES(2,'Otra','Recepcionista','trazabilidad@test.local','test-only','00000001')
		RETURNING id`).Scan(&segundoUsuarioID))

	p := f.order(t)
	r := NewPedidoRepository(f.db)
	observacionRack := "Prendas colocadas en rack"
	_, err := r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{
		EstadoID: f.state(t, "Rackeado"), Observaciones: &observacionRack,
	}, f.userID)
	mustSprint3(t, err)
	_, err = r.UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{
		EstadoID: f.state(t, "Finalizado"),
	}, segundoUsuarioID)
	mustSprint3(t, err)

	historial, err := r.GetHistorialEstados(f.ctx, p.ID)
	mustSprint3(t, err)
	if len(historial) != 3 {
		t.Fatalf("se esperaban recepción y dos cambios, se obtuvo %+v", historial)
	}
	esperados := []struct {
		estado    string
		usuarioID int
	}{
		{estado: "Recibido", usuarioID: f.userID},
		{estado: "Rackeado", usuarioID: f.userID},
		{estado: "Finalizado", usuarioID: segundoUsuarioID},
	}
	for i, esperado := range esperados {
		movimiento := historial[i]
		if movimiento.Estado.Nombre != esperado.estado || movimiento.UsuarioID != esperado.usuarioID || movimiento.Usuario.ID != esperado.usuarioID {
			t.Fatalf("movimiento %d sin estado/usuario esperado: %+v", i, movimiento)
		}
		if movimiento.FechaCambio.IsZero() || movimiento.CreatedAt.IsZero() {
			t.Fatalf("movimiento %d sin fecha y hora: %+v", i, movimiento)
		}
		if i > 0 && movimiento.FechaCambio.Before(historial[i-1].FechaCambio) {
			t.Fatalf("historial fuera de orden cronológico: %+v", historial)
		}
	}
	if historial[1].Observaciones == nil || *historial[1].Observaciones != observacionRack {
		t.Fatalf("observación de transición no persistida: %+v", historial[1])
	}
}
