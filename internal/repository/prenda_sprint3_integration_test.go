package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSprint3ServiciosTotalYPersistencia(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t, 2, 3)
	r := NewPrendaRepository(f.db)
	f.total(t, p.ID, 51.25)
	relation, err := r.AddServicio(f.ctx, p.Prendas[0].ID, f.services[1])
	mustSprint3(t, err)
	if relation.PrecioAplicado != 5.50 {
		t.Fatalf("precio debe provenir del catalogo: %+v", relation)
	}
	var price float64
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT precio_aplicado FROM prenda_servicios WHERE id=$1`, relation.ID).Scan(&price))
	if price != 5.50 {
		t.Fatalf("precio persistido=%v", price)
	}
	f.total(t, p.ID, 62.25)
	_, err = f.db.Exec(f.ctx, `UPDATE servicios SET precio_base=999 WHERE id=$1`, f.services[1])
	mustSprint3(t, err)
	// El precio aplicado es una instantanea: cambiar el catalogo no cambia el pedido.
	mustSprint3(t, r.RemoveServicio(f.ctx, p.Prendas[0].ID, f.services[0]))
	f.total(t, p.ID, 41.75)
	mustSprint3(t, r.RemoveServicio(f.ctx, p.Prendas[0].ID, f.services[1]))
	f.total(t, p.ID, 30.75)
	mustSprint3(t, r.RemoveServicio(f.ctx, p.Prendas[1].ID, f.services[0]))
	f.total(t, p.ID, 0)
}

func TestSprint3ServiciosRechazosNoAlteranTotal(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	r := NewPrendaRepository(f.db)
	for _, tc := range []struct {
		name             string
		prenda, servicio int
		want             error
	}{
		{"duplicado", p.Prendas[0].ID, f.services[0], ErrPrendaServicioYaAsociado},
		{"servicio inexistente", p.Prendas[0].ID, 2147483647, ErrServicioNoEncontrado},
		{"prenda inexistente", 2147483647, f.services[1], ErrPrendaNoEncontrada},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := r.AddServicio(f.ctx, tc.prenda, tc.servicio)
			if !errors.Is(err, tc.want) {
				t.Fatalf("esperado %v, obtenido %v", tc.want, err)
			}
			f.total(t, p.ID, 20.50)
		})
	}
	_, err := f.db.Exec(f.ctx, `UPDATE servicios SET activo=FALSE WHERE id=$1`, f.services[1])
	mustSprint3(t, err)
	_, err = r.AddServicio(f.ctx, p.Prendas[0].ID, f.services[1])
	if !errors.Is(err, ErrServicioNoEncontrado) {
		t.Fatalf("servicio inactivo: %v", err)
	}
	err = r.RemoveServicio(f.ctx, p.Prendas[0].ID, f.services[2])
	if !errors.Is(err, ErrPrendaServicioNoEncontrado) {
		t.Fatalf("relacion inexistente: %v", err)
	}
	f.total(t, p.ID, 20.50)
	var count int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM prenda_servicios WHERE prenda_id=$1`, p.Prendas[0].ID).Scan(&count))
	if count != 1 {
		t.Fatalf("relaciones tras rechazos=%d", count)
	}
}

func TestSprint3ServiciosRollbackSiFallaRecalculo(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	_, err := f.db.Exec(f.ctx, `CREATE FUNCTION fallar_total() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'fallo inducido de total'; END $$;
	 CREATE TRIGGER test_total BEFORE UPDATE ON pedidos FOR EACH ROW EXECUTE FUNCTION fallar_total()`)
	mustSprint3(t, err)
	r := NewPrendaRepository(f.db)
	if _, err = r.AddServicio(f.ctx, p.Prendas[0].ID, f.services[1]); err == nil {
		t.Fatal("esperaba fallo de recalculo")
	}
	if err = r.RemoveServicio(f.ctx, p.Prendas[0].ID, f.services[0]); err == nil {
		t.Fatal("esperaba fallo de recalculo al quitar")
	}
	var count int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM prenda_servicios WHERE prenda_id=$1 AND servicio_id=$2`, p.Prendas[0].ID, f.services[0]).Scan(&count))
	if count != 1 {
		t.Fatal("rollback no restauro servicio quitado")
	}
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM prenda_servicios WHERE prenda_id=$1`, p.Prendas[0].ID).Scan(&count))
	if count != 1 {
		t.Fatal("rollback dejo servicio agregado")
	}
	f.total(t, p.ID, 20.50)
}

func TestSprint3CrearPedidoRollbackPorServicioInexistente(t *testing.T) {
	f := newSprint3Fixture(t)
	_, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{
		{TipoPrendaID: f.typeID, Cantidad: 1, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}, {ServicioID: 2147483647}}},
	}}, f.userID)
	if !errors.Is(err, ErrServicioNoEncontrado) {
		t.Fatalf("esperaba servicio inexistente: %v", err)
	}
	for _, table := range []string{"pedidos", "prendas", "prenda_servicios"} {
		var count int
		mustSprint3(t, f.db.QueryRow(f.ctx, "SELECT COUNT(*) FROM "+table).Scan(&count))
		if count != 0 {
			t.Fatalf("rollback dejo %d registros en %s", count, table)
		}
	}
}

func TestSprint3EntregadoNoPermiteCambiarPrendasOServicios(t *testing.T) {
	f := newSprint3Fixture(t)
	p := f.order(t)
	_, err := NewPedidoRepository(f.db).UpdateEstado(f.ctx, p.ID, &models.PedidoEstadoUpdateRequest{EstadoID: f.state(t, "Entregado")}, f.userID)
	mustSprint3(t, err)
	r := NewPrendaRepository(f.db)
	_, err = r.AddServicio(f.ctx, p.Prendas[0].ID, f.services[1])
	if !errors.Is(err, ErrPedidoEstadoFinalizado) {
		t.Fatalf("agregar servicio a entregado: %v", err)
	}
	err = r.RemoveServicio(f.ctx, p.Prendas[0].ID, f.services[0])
	if !errors.Is(err, ErrPedidoEstadoFinalizado) {
		t.Fatalf("quitar servicio a entregado: %v", err)
	}
	_, err = r.CreateMany(f.ctx, p.ID, []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 1}})
	if !errors.Is(err, ErrPedidoEstadoFinalizado) {
		t.Fatalf("agregar prenda a entregado: %v", err)
	}
	f.total(t, p.ID, 20.50)
}
