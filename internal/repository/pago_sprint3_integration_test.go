package repository

import (
	"errors"
	"math"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSprint3PagoRegistraActualizaSaldoYApareceEnHistorial(t *testing.T) {
	f := newSprint3Fixture(t)
	pedido := f.order(t, 2)
	var metodoID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO metodos_pago(nombre) VALUES('Tarjeta de prueba') RETURNING id`).Scan(&metodoID))
	referencia := "POS-456"
	repo := NewPagoRepository(f.db)

	pago, err := repo.Create(f.ctx, pedido.ID, &models.PagoCreateRequest{MetodoPagoID: metodoID, Monto: 8.50, Referencia: &referencia}, f.userID)
	mustSprint3(t, err)
	if pago.ID == 0 || pago.MetodoPago.Nombre != "Tarjeta de prueba" {
		t.Fatalf("pago inesperado: %+v", pago)
	}

	saldo, err := repo.GetSaldo(f.ctx, pedido.ID)
	mustSprint3(t, err)
	if math.Abs(saldo.Total-20.50) > 0.00001 || math.Abs(saldo.TotalPagado-8.50) > 0.00001 || math.Abs(saldo.SaldoPendiente-12) > 0.00001 {
		t.Fatalf("saldo inesperado: %+v", saldo)
	}

	historial, err := repo.ListByPedido(f.ctx, pedido.ID)
	mustSprint3(t, err)
	if len(historial) != 1 || historial[0].Referencia == nil || *historial[0].Referencia != referencia {
		t.Fatalf("historial inesperado: %+v", historial)
	}
	if historial[0].Usuario == nil || historial[0].Usuario.ID != f.userID || historial[0].Usuario.Nombre != "Usuario" {
		t.Fatalf("usuario del historial inesperado: %+v", historial[0].Usuario)
	}
}

func TestSprint3PagoRechazaSobrepagoYPedidoCancelado(t *testing.T) {
	f := newSprint3Fixture(t)
	pedido := f.order(t, 1)
	var metodoID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO metodos_pago(nombre) VALUES('Efectivo de prueba') RETURNING id`).Scan(&metodoID))
	repo := NewPagoRepository(f.db)

	_, err := repo.Create(f.ctx, pedido.ID, &models.PagoCreateRequest{MetodoPagoID: metodoID, Monto: 10.26}, f.userID)
	if !errors.Is(err, ErrPagoExcedeSaldo) {
		t.Fatalf("sobrepago: error = %v", err)
	}

	_, err = NewPedidoRepository(f.db).Cancelar(f.ctx, pedido.ID, f.userID)
	mustSprint3(t, err)
	_, err = repo.Create(f.ctx, pedido.ID, &models.PagoCreateRequest{MetodoPagoID: metodoID, Monto: 5}, f.userID)
	if !errors.Is(err, ErrPagoPedidoCancelado) {
		t.Fatalf("pago de pedido cancelado: error = %v", err)
	}
	historial, err := repo.ListByPedido(f.ctx, pedido.ID)
	mustSprint3(t, err)
	if len(historial) != 0 {
		t.Fatalf("historial del pedido cancelado = %+v", historial)
	}
}
