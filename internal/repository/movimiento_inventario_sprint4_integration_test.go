package repository

import (
	"errors"
	"math"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func sprint4Input(t *testing.T, f *sprint3Fixture, name string, stock, minimum float64) *models.Insumo {
	t.Helper()
	input, err := NewInsumoRepository(f.db).Create(f.ctx, &models.InsumoCreateRequest{
		Nombre:       name,
		UnidadMedida: "L",
		StockActual:  stock,
		StockMinimo:  minimum,
	})
	mustSprint3(t, err)
	return input
}

func sprint4Stock(t *testing.T, f *sprint3Fixture, inputID int) float64 {
	t.Helper()
	var stock float64
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT stock_actual FROM insumos WHERE id=$1`, inputID).Scan(&stock))
	return stock
}

func sprint4MovementCount(t *testing.T, f *sprint3Fixture, inputID int) int {
	t.Helper()
	var count int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM movimientos_inventario WHERE insumo_id=$1`, inputID).Scan(&count))
	return count
}

func assertSprint4Stock(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.00001 {
		t.Fatalf("stock=%v, expected %v", got, want)
	}
}

func TestSprint4MovimientoEntradaYSalidaActualizanStockYPersisten(t *testing.T) {
	f := newSprint3Fixture(t)
	input := sprint4Input(t, f, "Detergente Sprint 4", 10, 5)
	repo := NewMovimientoInventarioRepository(f.db)
	entryReason := "Compra"

	entry, err := repo.Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: input.ID, TipoMovimiento: "entrada", Cantidad: 5, Motivo: &entryReason,
	}, f.userID)
	mustSprint3(t, err)
	if entry.ID == 0 || entry.UsuarioID != f.userID || entry.TipoMovimiento != "entrada" || entry.Motivo == nil || *entry.Motivo != entryReason {
		t.Fatalf("unexpected persisted entry: %+v", entry)
	}
	assertSprint4Stock(t, sprint4Stock(t, f, input.ID), 15)

	exit, err := repo.Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: input.ID, TipoMovimiento: "salida", Cantidad: 4,
	}, f.userID)
	mustSprint3(t, err)
	if exit.ID == 0 || exit.TipoMovimiento != "salida" || exit.Cantidad != 4 {
		t.Fatalf("unexpected persisted exit: %+v", exit)
	}
	assertSprint4Stock(t, sprint4Stock(t, f, input.ID), 11)
	if count := sprint4MovementCount(t, f, input.ID); count != 2 {
		t.Fatalf("movement count=%d, expected 2", count)
	}
}

func TestSprint4MovimientoRechazaSalidaExcesivaSinCambios(t *testing.T) {
	f := newSprint3Fixture(t)
	input := sprint4Input(t, f, "Desmanchador Sprint 4", 3, 1)
	repo := NewMovimientoInventarioRepository(f.db)

	_, err := repo.Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: input.ID, TipoMovimiento: "salida", Cantidad: 5,
	}, f.userID)
	if !errors.Is(err, ErrStockInsuficiente) {
		t.Fatalf("error=%v, expected ErrStockInsuficiente", err)
	}
	assertSprint4Stock(t, sprint4Stock(t, f, input.ID), 3)
	if count := sprint4MovementCount(t, f, input.ID); count != 0 {
		t.Fatalf("movement count=%d, expected 0", count)
	}
}

func TestSprint4MovimientoRechazaInsumoYUsuarioInexistentes(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewMovimientoInventarioRepository(f.db)

	_, err := repo.Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: 999999, TipoMovimiento: "entrada", Cantidad: 1,
	}, f.userID)
	if !errors.Is(err, ErrInsumoNoEncontrado) {
		t.Fatalf("missing input error=%v", err)
	}

	input := sprint4Input(t, f, "Suavizante Sprint 4", 8, 2)
	_, err = repo.Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: input.ID, TipoMovimiento: "entrada", Cantidad: 2,
	}, 999999)
	if !errors.Is(err, ErrUsuarioNoEncontrado) {
		t.Fatalf("missing user error=%v", err)
	}
	assertSprint4Stock(t, sprint4Stock(t, f, input.ID), 8)
	if count := sprint4MovementCount(t, f, input.ID); count != 0 {
		t.Fatalf("movement count=%d, expected 0", count)
	}
}

func TestSprint4MovimientoRevierteStockSiFallaPersistencia(t *testing.T) {
	f := newSprint3Fixture(t)
	input := sprint4Input(t, f, "Rollback Sprint 4", 10, 2)
	_, err := f.db.Exec(f.ctx, `
		CREATE FUNCTION sprint4_reject_movement() RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced movement insert failure';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER sprint4_reject_movement_trigger
		BEFORE INSERT ON movimientos_inventario
		FOR EACH ROW EXECUTE FUNCTION sprint4_reject_movement();`)
	mustSprint3(t, err)

	_, err = NewMovimientoInventarioRepository(f.db).Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: input.ID, TipoMovimiento: "entrada", Cantidad: 5,
	}, f.userID)
	if err == nil {
		t.Fatal("expected forced persistence error")
	}
	assertSprint4Stock(t, sprint4Stock(t, f, input.ID), 10)
	if count := sprint4MovementCount(t, f, input.ID); count != 0 {
		t.Fatalf("movement count=%d, expected 0", count)
	}
}

func TestSprint4MovimientoHaceTransicionAStockMinimo(t *testing.T) {
	f := newSprint3Fixture(t)
	input := sprint4Input(t, f, "Alerta Sprint 4", 6, 5)
	_, err := NewMovimientoInventarioRepository(f.db).Create(f.ctx, &models.MovimientoInventarioCreateRequest{
		InsumoID: input.ID, TipoMovimiento: "salida", Cantidad: 2,
	}, f.userID)
	mustSprint3(t, err)

	lowStock, err := NewInsumoRepository(f.db).ListLowStock(f.ctx)
	mustSprint3(t, err)
	if len(lowStock) != 1 || lowStock[0].ID != input.ID {
		t.Fatalf("low stock list=%+v, expected input %d", lowStock, input.ID)
	}
	assertSprint4Stock(t, lowStock[0].StockActual, 4)
}
