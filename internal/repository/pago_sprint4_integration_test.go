package repository

import (
	"errors"
	"math"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func sprint4PaymentMethod(t *testing.T, f *sprint3Fixture, name string) int {
	t.Helper()
	var id int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO metodos_pago(nombre) VALUES($1) RETURNING id`, name).Scan(&id))
	return id
}

func assertSprint4Amount(t *testing.T, got, want float64, label string) {
	t.Helper()
	if math.Abs(got-want) > 0.00001 {
		t.Fatalf("%s=%v, expected %v", label, got, want)
	}
}

func TestSprint4PagoExactoYMultiplesPagosCalculanSaldo(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPagoRepository(f.db)
	methodID := sprint4PaymentMethod(t, f, "Pago Sprint 4")

	exactOrder := f.order(t, 1)
	_, err := repo.Create(f.ctx, exactOrder.ID, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 10.25}, f.userID)
	mustSprint3(t, err)
	exactBalance, err := repo.GetSaldo(f.ctx, exactOrder.ID)
	mustSprint3(t, err)
	assertSprint4Amount(t, exactBalance.TotalPagado, 10.25, "exact total paid")
	assertSprint4Amount(t, exactBalance.SaldoPendiente, 0, "exact balance")

	multipleOrder := f.order(t, 2)
	firstReference := "S4-1"
	secondReference := "S4-2"
	first, err := repo.Create(f.ctx, multipleOrder.ID, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 7.25, Referencia: &firstReference}, f.userID)
	mustSprint3(t, err)
	second, err := repo.Create(f.ctx, multipleOrder.ID, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 13.25, Referencia: &secondReference}, f.userID)
	mustSprint3(t, err)

	balance, err := repo.GetSaldo(f.ctx, multipleOrder.ID)
	mustSprint3(t, err)
	assertSprint4Amount(t, balance.Total, 20.50, "total")
	assertSprint4Amount(t, balance.TotalPagado, 20.50, "total paid")
	assertSprint4Amount(t, balance.SaldoPendiente, 0, "balance")

	history, err := repo.ListByPedido(f.ctx, multipleOrder.ID)
	mustSprint3(t, err)
	if len(history) != 2 {
		t.Fatalf("history length=%d, expected 2", len(history))
	}
	if history[0].ID != first.ID || history[1].ID != second.ID || history[0].ID >= history[1].ID {
		t.Fatalf("history is not chronological: %+v", history)
	}
	if history[0].Referencia == nil || *history[0].Referencia != firstReference || history[1].Referencia == nil || *history[1].Referencia != secondReference {
		t.Fatalf("history references are incomplete: %+v", history)
	}

	_, err = repo.Create(f.ctx, multipleOrder.ID, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 1}, f.userID)
	if !errors.Is(err, ErrPedidoSinSaldo) {
		t.Fatalf("payment after exact balance error=%v, expected ErrPedidoSinSaldo", err)
	}
}

func TestSprint4PagoRechazaReferenciasInexistentes(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPagoRepository(f.db)
	methodID := sprint4PaymentMethod(t, f, "Valid method")
	order := f.order(t, 1)

	_, err := repo.Create(f.ctx, 999999, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 1}, f.userID)
	if !errors.Is(err, ErrPedidoNoEncontrado) {
		t.Fatalf("missing order error=%v", err)
	}
	_, err = repo.GetSaldo(f.ctx, 999999)
	if !errors.Is(err, ErrPedidoNoEncontrado) {
		t.Fatalf("missing balance order error=%v", err)
	}
	_, err = repo.ListByPedido(f.ctx, 999999)
	if !errors.Is(err, ErrPedidoNoEncontrado) {
		t.Fatalf("missing history order error=%v", err)
	}

	_, err = repo.Create(f.ctx, order.ID, &models.PagoCreateRequest{MetodoPagoID: 999999, Monto: 1}, f.userID)
	if !errors.Is(err, ErrMetodoPagoNoEncontrado) {
		t.Fatalf("missing payment method error=%v", err)
	}
	_, err = repo.Create(f.ctx, order.ID, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 1}, 999999)
	if !errors.Is(err, ErrUsuarioNoEncontrado) {
		t.Fatalf("missing user error=%v", err)
	}

	balance, err := repo.GetSaldo(f.ctx, order.ID)
	mustSprint3(t, err)
	assertSprint4Amount(t, balance.TotalPagado, 0, "total paid after rejected payments")
}

func TestSprint4PagosConcurrentesNoExcedenSaldo(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPagoRepository(f.db)
	methodID := sprint4PaymentMethod(t, f, "Concurrent method")
	order := f.order(t, 1)
	errorsChannel := make(chan error, 2)

	for range 2 {
		go func() {
			_, err := repo.Create(f.ctx, order.ID, &models.PagoCreateRequest{MetodoPagoID: methodID, Monto: 6}, f.userID)
			errorsChannel <- err
		}()
	}

	successes := 0
	overpayments := 0
	for range 2 {
		err := <-errorsChannel
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrPagoExcedeSaldo):
			overpayments++
		default:
			t.Fatalf("unexpected concurrent payment error: %v", err)
		}
	}
	if successes != 1 || overpayments != 1 {
		t.Fatalf("successes=%d overpayments=%d, expected 1 and 1", successes, overpayments)
	}
	balance, err := repo.GetSaldo(f.ctx, order.ID)
	mustSprint3(t, err)
	assertSprint4Amount(t, balance.TotalPagado, 6, "concurrent total paid")
	assertSprint4Amount(t, balance.SaldoPendiente, 4.25, "concurrent balance")
}
