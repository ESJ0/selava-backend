package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestPagoClienteAislaHistorialYSaldoPorPropietario(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPagoRepository(f.db)
	metodoID := sprint4PaymentMethod(t, f, "Transferencia cliente")
	pedidoPropio := f.order(t, 2)
	_, err := repo.Create(f.ctx, pedidoPropio.ID, &models.PagoCreateRequest{MetodoPagoID: metodoID, Monto: 8.50}, f.userID)
	mustSprint3(t, err)

	var otroClienteID, pedidoAjenoID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `
		INSERT INTO clientes(nombre, apellido, telefono)
		VALUES ('Otro', 'Cliente', '11111111') RETURNING id`).Scan(&otroClienteID))
	mustSprint3(t, f.db.QueryRow(f.ctx, `
		INSERT INTO pedidos(cliente_id, usuario_id, estado_actual_id, total)
		SELECT $1, $2, id, 40 FROM estados_pedido WHERE nombre = 'Recibido'
		RETURNING id`, otroClienteID, f.userID).Scan(&pedidoAjenoID))

	pagos, err := repo.ListByPedidoCliente(f.ctx, pedidoPropio.ID, f.clientID)
	mustSprint3(t, err)
	if len(pagos) != 1 || pagos[0].Monto != 8.50 {
		t.Fatalf("historial propio inesperado: %+v", pagos)
	}
	saldo, err := repo.GetSaldoCliente(f.ctx, pedidoPropio.ID, f.clientID)
	mustSprint3(t, err)
	if saldo.TotalPagado != 8.50 || saldo.SaldoPendiente != 12 {
		t.Fatalf("saldo propio inesperado: %+v", saldo)
	}

	for _, call := range []func() error{
		func() error { _, err := repo.ListByPedidoCliente(f.ctx, pedidoAjenoID, f.clientID); return err },
		func() error { _, err := repo.GetSaldoCliente(f.ctx, pedidoAjenoID, f.clientID); return err },
		func() error { _, err := repo.ListByPedidoCliente(f.ctx, pedidoPropio.ID, otroClienteID); return err },
		func() error { _, err := repo.GetSaldoCliente(f.ctx, pedidoPropio.ID, otroClienteID); return err },
	} {
		if err := call(); !errors.Is(err, ErrPedidoNoEncontrado) {
			t.Fatalf("pedido ajeno error=%v, esperado ErrPedidoNoEncontrado", err)
		}
	}
}

func TestPagoClientePedidoPropioSinPagosDevuelveArregloVacioYSaldoCompleto(t *testing.T) {
	f := newSprint3Fixture(t)
	pedido := f.order(t, 1)
	repo := NewPagoRepository(f.db)

	pagos, err := repo.ListByPedidoCliente(f.ctx, pedido.ID, f.clientID)
	mustSprint3(t, err)
	if pagos == nil || len(pagos) != 0 {
		t.Fatalf("historial vacio inesperado: %#v", pagos)
	}
	saldo, err := repo.GetSaldoCliente(f.ctx, pedido.ID, f.clientID)
	mustSprint3(t, err)
	if saldo.TotalPagado != 0 || saldo.SaldoPendiente != saldo.Total {
		t.Fatalf("saldo sin pagos inesperado: %+v", saldo)
	}
}
