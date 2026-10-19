package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSEL107TipoPrendaCRUDYRestricciones(t *testing.T) {
	f := newSprint3Fixture(t)
	r := NewTipoPrendaRepository(f.db)
	precio := 25.50
	tipo, err := r.Create(f.ctx, &models.TipoPrendaCreateRequest{Nombre: "Pantalon", Precio: &precio})
	mustSprint3(t, err)
	if tipo.Precio == nil || *tipo.Precio != precio {
		t.Fatal(tipo)
	}
	precio = 30.25
	tipo, err = r.Update(f.ctx, tipo.ID, &models.TipoPrendaUpdateRequest{Precio: &precio})
	mustSprint3(t, err)
	tipo, err = r.GetByID(f.ctx, tipo.ID)
	mustSprint3(t, err)
	if tipo.Precio == nil || *tipo.Precio != 30.25 {
		t.Fatal(tipo)
	}
	tipos, err := r.List(f.ctx)
	mustSprint3(t, err)
	if len(tipos) != 2 || tipos[1].Precio == nil || *tipos[1].Precio != 30.25 {
		t.Fatal(tipos)
	}
	for _, invalid := range []string{"NULL", "0", "-1", "1.001", "100000000", "'NaN'", "'Infinity'"} {
		_, err = f.db.Exec(f.ctx, "INSERT INTO tipos_prenda(nombre,precio) VALUES('Invalido',"+invalid+")")
		if err == nil {
			t.Fatalf("precio invalido aceptado: %s", invalid)
		}
	}
	_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=NULL WHERE id=$1`, tipo.ID)
	if err == nil {
		t.Fatal("permitio borrar precio configurado")
	}
	public, err := r.ListActiveWithPrice(f.ctx)
	mustSprint3(t, err)
	if len(public) != 2 || public[1].Precio != 30.25 {
		t.Fatal(public)
	}
	mustSprint3(t, r.Delete(f.ctx, tipo.ID))
	public, err = r.ListActiveWithPrice(f.ctx)
	mustSprint3(t, err)
	if len(public) != 1 || public[0].ID != f.typeID {
		t.Fatal(public)
	}
	sr := NewServicioRepository(f.db)
	servicio, err := sr.Create(f.ctx, &models.ServicioCreateRequest{Nombre: "Sin tarifa"})
	mustSprint3(t, err)
	_, err = sr.GetByID(f.ctx, servicio.ID)
	mustSprint3(t, err)
	_, err = sr.ListActive(f.ctx)
	mustSprint3(t, err)
}

func TestSEL107VariosServiciosSinRecargoYPrecioCongelado(t *testing.T) {
	f := newSprint3Fixture(t)
	_, err := f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=25 WHERE id=$1`, f.typeID)
	mustSprint3(t, err)
	p, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 2, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}, {ServicioID: f.services[1]}}}}}, f.userID)
	mustSprint3(t, err)
	f.total(t, p.ID, 50)
	if p.Prendas[0].PrecioAplicado == nil || *p.Prendas[0].PrecioAplicado != 25 {
		t.Fatal(p.Prendas)
	}
	for _, s := range p.Prendas[0].Servicios {
		if s.PrecioAplicado != 0 {
			t.Fatal(s)
		}
	}
	_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=30 WHERE id=$1;`, f.typeID)
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, `UPDATE servicios SET precio_base=999`)
	mustSprint3(t, err)
	f.total(t, p.ID, 50)
	r := NewPrendaRepository(f.db)
	_, err = r.AddServicio(f.ctx, p.Prendas[0].ID, f.services[2])
	mustSprint3(t, err)
	f.total(t, p.ID, 50)
	mustSprint3(t, r.RemoveServicio(f.ctx, p.Prendas[0].ID, f.services[0]))
	f.total(t, p.ID, 50)
	detalle, err := NewPedidoRepository(f.db).GetDetalle(f.ctx, p.ID)
	mustSprint3(t, err)
	if *detalle.Prendas[0].PrecioAplicado != 25 || *detalle.Prendas[0].TipoPrenda.Precio != 30 {
		t.Fatal(detalle.Prendas)
	}
	newer := f.order(t, 1)
	f.total(t, newer.ID, 30)
	prendas, err := r.CreateMany(f.ctx, p.ID, []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 1}})
	mustSprint3(t, err)
	if *prendas[0].PrecioAplicado != 30 {
		t.Fatal(prendas)
	}
	f.total(t, p.ID, 80)
}

func TestSEL107MigracionPreservaPedidosAnteriores(t *testing.T) {
	f := newSprint3Fixture(t, true)
	var pedidoID, prendaID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO pedidos(cliente_id,usuario_id,estado_actual_id,total) VALUES($1,$2,$3,31.50) RETURNING id`, f.clientID, f.userID, f.state(t, "Recibido")).Scan(&pedidoID))
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO prendas(pedido_id,tipo_prenda_id,cantidad) VALUES($1,$2,2) RETURNING id`, pedidoID, f.typeID).Scan(&prendaID))
	_, err := f.db.Exec(f.ctx, `INSERT INTO prenda_servicios(prenda_id,servicio_id,precio_aplicado) VALUES($1,$2,10.25),($1,$3,5.50)`, prendaID, f.services[0], f.services[1])
	mustSprint3(t, err)
	migration, err := os.ReadFile("../../migrations/023_garment_type_pricing.up.sql")
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, string(migration))
	mustSprint3(t, err)
	chargeModes, err := os.ReadFile("../../migrations/024_garment_charge_modes.up.sql")
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, string(chargeModes))
	mustSprint3(t, err)
	f.total(t, pedidoID, 31.50)
	public, err := NewTipoPrendaRepository(f.db).ListActiveWithPrice(f.ctx)
	mustSprint3(t, err)
	if len(public) != 0 {
		t.Fatal("publico tipo sin precio", public)
	}
	_, err = NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 1}}}, f.userID)
	if !errors.Is(err, ErrTipoPrendaSinPrecio) {
		t.Fatalf("sin precio: %v", err)
	}
	_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=99 WHERE id=$1`, f.typeID)
	mustSprint3(t, err)
	// Recalcular sigue utilizando exclusivamente las instantaneas anteriores.
	tx, err := f.db.Begin(f.ctx)
	mustSprint3(t, err)
	_, _, err = recalcularTotalPedido(f.ctx, tx, pedidoID)
	mustSprint3(t, err)
	mustSprint3(t, tx.Commit(f.ctx))
	f.total(t, pedidoID, 31.50)
	detalle, err := NewPedidoRepository(f.db).GetDetalle(f.ctx, pedidoID)
	mustSprint3(t, err)
	if detalle.Prendas[0].PrecioAplicado != nil || detalle.Prendas[0].Servicios[0].PrecioAplicado != 10.25 {
		t.Fatal(detalle.Prendas)
	}
	if detalle.Prendas[0].Subtotal == nil || *detalle.Prendas[0].Subtotal != 31.50 {
		t.Fatal("subtotal historico alterado", detalle.Prendas)
	}
	nuevo := f.order(t, 1)
	f.total(t, nuevo.ID, 99)
	f.total(t, pedidoID, 31.50)
	down, err := os.ReadFile("../../migrations/023_garment_type_pricing.down.sql")
	mustSprint3(t, err)
	if _, err = f.db.Exec(f.ctx, string(down)); err == nil {
		t.Fatal("rollback destructivo permitido")
	}
	f.total(t, nuevo.ID, 99)
}
