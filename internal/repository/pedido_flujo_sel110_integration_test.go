package repository

import (
	"errors"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSEL110PedidoRecorreFlujoCompletoHastaFinalizado(t *testing.T) {
	f := newSprint3Fixture(t)
	servicios := make(map[string]int)
	for _, nombre := range []string{"Lavado", "Planchado"} {
		var id int
		mustSprint3(t, f.db.QueryRow(f.ctx,
			`INSERT INTO servicios(nombre, precio_base) VALUES($1, 1) RETURNING id`, nombre,
		).Scan(&id))
		servicios[nombre] = id
	}

	r := NewPedidoRepository(f.db)
	pedido, err := r.Create(f.ctx, &models.PedidoCreateRequest{
		ClienteID: f.clientID,
		Prendas: []models.PrendaCreateRequest{{
			TipoPrendaID: f.typeID,
			Cantidad:     1,
			Servicios: []models.PrendaServicioCreateRequest{
				{ServicioID: servicios["Lavado"]},
				{ServicioID: servicios["Planchado"]},
			},
		}},
	}, f.userID)
	mustSprint3(t, err)

	assertFlow := func(actual string, progress int, next string) {
		t.Helper()
		detail, detailErr := r.GetDetalle(f.ctx, pedido.ID)
		mustSprint3(t, detailErr)
		if detail.EstadoActual.Nombre != actual || detail.FlujoOperativo.ProgresoPorcentaje != progress {
			t.Fatalf("estado/progreso inesperado: estado=%s progreso=%d flujo=%+v", detail.EstadoActual.Nombre, detail.FlujoOperativo.ProgresoPorcentaje, detail.FlujoOperativo)
		}
		if next == "" {
			if detail.FlujoOperativo.SiguienteEstado != nil || detail.FechaEntregaReal == nil {
				t.Fatalf("pedido finalizado con siguiente etapa o sin fecha real: %+v", detail)
			}
			return
		}
		if detail.FlujoOperativo.SiguienteEstado == nil || detail.FlujoOperativo.SiguienteEstado.Nombre != next {
			t.Fatalf("siguiente etapa=%+v, esperada=%s", detail.FlujoOperativo.SiguienteEstado, next)
		}
	}

	assertFlow("Recibido", 20, "En lavado")
	for _, destino := range []string{"En planchado", "Rackeado", "Finalizado"} {
		_, err = r.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, destino)}, f.userID)
		if !errors.Is(err, ErrTransicionEstadoNoPermitida) {
			t.Fatalf("se permitió saltar de Recibido a %s: %v", destino, err)
		}
	}

	_, err = r.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "En lavado")}, f.userID)
	mustSprint3(t, err)
	assertFlow("En lavado", 40, "En planchado")

	for _, tc := range []struct {
		destino string
		want    error
	}{
		{destino: "Recibido", want: ErrPedidoEstadoRetrocedido},
		{destino: "En lavado", want: ErrTransicionEstadoNoPermitida},
		{destino: "Finalizado", want: ErrTransicionEstadoNoPermitida},
	} {
		_, got := r.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, tc.destino)}, f.userID)
		if !errors.Is(got, tc.want) {
			t.Fatalf("destino inválido %s: esperado=%v obtenido=%v", tc.destino, tc.want, got)
		}
	}

	for _, paso := range []struct {
		destino   string
		progress  int
		siguiente string
	}{
		{destino: "En planchado", progress: 60, siguiente: "Rackeado"},
		{destino: "Rackeado", progress: 80, siguiente: "Finalizado"},
		{destino: "Finalizado", progress: 100},
	} {
		_, err = r.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, paso.destino)}, f.userID)
		mustSprint3(t, err)
		assertFlow(paso.destino, paso.progress, paso.siguiente)
	}

	_, err = r.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Rackeado")}, f.userID)
	if !errors.Is(err, ErrPedidoEstadoFinalizado) {
		t.Fatalf("pedido finalizado aceptó otro cambio: %v", err)
	}

	historial, err := r.GetHistorialEstados(f.ctx, pedido.ID)
	mustSprint3(t, err)
	nombres := make([]string, 0, len(historial))
	for _, movimiento := range historial {
		nombres = append(nombres, movimiento.Estado.Nombre)
	}
	const esperado = "Recibido > En lavado > En planchado > Rackeado > Finalizado"
	if strings.Join(nombres, " > ") != esperado {
		t.Fatalf("historial=%s, esperado=%s", strings.Join(nombres, " > "), esperado)
	}
}
