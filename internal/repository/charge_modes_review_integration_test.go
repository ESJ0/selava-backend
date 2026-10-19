package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSEL108ReviewMigracionNoReinterpretaCanastaHistorica(t *testing.T) {
	f := newSprint3Fixture(t, true)
	sql, err := os.ReadFile("../../migrations/023_garment_type_pricing.up.sql")
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, string(sql))
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET nombre='Canasta',precio=10.25 WHERE id=$1`, f.typeID)
	mustSprint3(t, err)
	var orderID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO pedidos(cliente_id,usuario_id,estado_actual_id,total) VALUES($1,$2,$3,20.50) RETURNING id`, f.clientID, f.userID, f.state(t, "Recibido")).Scan(&orderID))
	_, err = f.db.Exec(f.ctx, `INSERT INTO prendas(pedido_id,tipo_prenda_id,cantidad,precio_aplicado) VALUES($1,$2,2,10.25)`, orderID, f.typeID)
	mustSprint3(t, err)
	sql, err = os.ReadFile("../../migrations/024_garment_charge_modes.up.sql")
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, string(sql))
	mustSprint3(t, err)
	detail, err := NewPedidoRepository(f.db).GetDetalle(f.ctx, orderID)
	mustSprint3(t, err)
	if detail.Prendas[0].ModalidadCobroAplicada != models.ModalidadCobroUnidad || detail.Prendas[0].TipoPrenda.ModalidadCobro != models.ModalidadCobroLibra {
		t.Fatal("historico reinterpretado", detail.Prendas)
	}
	f.total(t, orderID, 20.50)
	// Sin nuevas medidas, el down es valido y conserva el importe anterior.
	down, err := os.ReadFile("../../migrations/024_garment_charge_modes.down.sql")
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, string(down))
	mustSprint3(t, err)
	f.total(t, orderID, 20.50)
}

func TestSEL108ReviewRedondeoPorSubtotalSinRecargos(t *testing.T) {
	f := newSprint3Fixture(t)
	price := 1.01
	kind, err := NewTipoPrendaRepository(f.db).Create(f.ctx, &models.TipoPrendaCreateRequest{Nombre: "Medida", Precio: &price, ModalidadCobro: models.ModalidadCobroMetro})
	mustSprint3(t, err)
	garment := models.PrendaCreateRequest{TipoPrendaID: kind.ID, Cantidad: 0.5, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}, {ServicioID: f.services[1]}}}
	order, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{garment, garment}}, f.userID)
	mustSprint3(t, err)
	f.total(t, order.ID, 1.02)
	detail, err := NewPedidoRepository(f.db).GetDetalle(f.ctx, order.ID)
	mustSprint3(t, err)
	for _, item := range detail.Prendas {
		if item.Subtotal == nil || *item.Subtotal != 0.51 {
			t.Fatalf("subtotal sin redondeo del servidor: %+v", item)
		}
	}
	for _, item := range order.Prendas {
		for _, service := range item.Servicios {
			if service.PrecioAplicado != 0 {
				t.Fatal(service)
			}
		}
	}
	// Ni la tarifa del servicio ni cambiar asociaciones alteran el cobro.
	_, err = f.db.Exec(f.ctx, `UPDATE servicios SET precio_base=999`)
	mustSprint3(t, err)
	_, err = NewPrendaRepository(f.db).AddServicio(f.ctx, order.Prendas[0].ID, f.services[2])
	mustSprint3(t, err)
	f.total(t, order.ID, 1.02)
	mustSprint3(t, NewPrendaRepository(f.db).RemoveServicio(f.ctx, order.Prendas[0].ID, f.services[0]))
	f.total(t, order.ID, 1.02)
}

func TestSEL108ReviewDownProtegeLibrasEnterasYRollbackDeImporte(t *testing.T) {
	f := newSprint3Fixture(t)
	price := 25.0
	kind, err := NewTipoPrendaRepository(f.db).Create(f.ctx, &models.TipoPrendaCreateRequest{Nombre: "Canasta", Precio: &price, ModalidadCobro: models.ModalidadCobroLibra})
	mustSprint3(t, err)
	order, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{TipoPrendaID: kind.ID, Cantidad: 2}}}, f.userID)
	mustSprint3(t, err)
	down, err := os.ReadFile("../../migrations/024_garment_charge_modes.down.sql")
	mustSprint3(t, err)
	if _, err = f.db.Exec(f.ctx, string(down)); err == nil {
		t.Fatal("down perdio modalidad LIBRA por ser cantidad entera")
	}
	huge := 99999999.99
	_, err = NewTipoPrendaRepository(f.db).Update(f.ctx, kind.ID, &models.TipoPrendaUpdateRequest{Precio: &huge})
	mustSprint3(t, err)
	_, err = NewPrendaRepository(f.db).CreateMany(f.ctx, order.ID, []models.PrendaCreateRequest{{TipoPrendaID: kind.ID, Cantidad: 2}})
	if !errors.Is(err, ErrTotalPedidoFueraDeRango) {
		t.Fatal("importe fuera de rango sin error de dominio", err)
	}
	f.total(t, order.ID, 50)
	var count int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM prendas WHERE pedido_id=$1`, order.ID).Scan(&count))
	if count != 1 {
		t.Fatal("alta parcial tras rollback", count)
	}
}
