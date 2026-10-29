package repository

import (
	"errors"
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestNotificacionesAisladasPorCliente(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewNotificacionRepository(f.db)

	var otroClienteID int
	mustSprint3(t, f.db.QueryRow(f.ctx,
		`INSERT INTO clientes(nombre, apellido, telefono) VALUES('Otro','Cliente','11111111') RETURNING id`,
	).Scan(&otroClienteID))
	propia, err := repo.Create(f.ctx, f.clientID, nil, "Aviso propio")
	mustSprint3(t, err)
	ajena, err := repo.Create(f.ctx, otroClienteID, nil, "Aviso ajeno")
	mustSprint3(t, err)

	listado, err := repo.ListByCliente(f.ctx, f.clientID)
	mustSprint3(t, err)
	if len(listado.Notificaciones) != 1 || listado.Notificaciones[0].ID != propia.ID || listado.NoLeidas != 1 {
		t.Fatalf("listado propio inesperado: %+v", listado)
	}
	if _, err := repo.MarkReadByCliente(f.ctx, ajena.ID, f.clientID); !errors.Is(err, ErrNotificacionNoEncontrada) {
		t.Fatalf("marcar notificacion ajena error=%v", err)
	}
	if err := repo.DeleteByCliente(f.ctx, ajena.ID, f.clientID); !errors.Is(err, ErrNotificacionNoEncontrada) {
		t.Fatalf("eliminar notificacion ajena error=%v", err)
	}
	if _, err := repo.MarkReadByCliente(f.ctx, propia.ID, f.clientID); err != nil {
		t.Fatal(err)
	}
	listado, err = repo.ListByCliente(f.ctx, f.clientID)
	mustSprint3(t, err)
	if listado.NoLeidas != 0 || !listado.Notificaciones[0].Leida {
		t.Fatalf("estado leido no persistido: %+v", listado)
	}
}

func TestCambioEstadoGeneraNotificacionYMensajeEspecialAlFinalizar(t *testing.T) {
	f := newSprint3Fixture(t)
	pedido := f.order(t)
	pedidos := NewPedidoRepository(f.db)
	notificaciones := NewNotificacionRepository(f.db)

	if _, err := pedidos.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{
		EstadoID: f.state(t, "Rackeado"),
	}, f.userID); err != nil {
		t.Fatal(err)
	}
	listado, err := notificaciones.ListByCliente(f.ctx, f.clientID)
	mustSprint3(t, err)
	if len(listado.Notificaciones) != 1 || !strings.Contains(listado.Notificaciones[0].Mensaje, "Rackeado") {
		t.Fatalf("notificacion de cambio inesperada: %+v", listado)
	}

	if _, err := pedidos.UpdateEstado(f.ctx, pedido.ID, &models.PedidoEstadoUpdateRequest{
		EstadoID: f.state(t, "Finalizado"),
	}, f.userID); err != nil {
		t.Fatal(err)
	}
	listado, err = notificaciones.ListByCliente(f.ctx, f.clientID)
	mustSprint3(t, err)
	if len(listado.Notificaciones) != 2 || !strings.Contains(strings.ToLower(listado.Notificaciones[0].Mensaje), "finalizado") {
		t.Fatalf("notificacion final inesperada: %+v", listado)
	}
}
