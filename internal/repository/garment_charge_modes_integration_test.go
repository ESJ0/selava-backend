package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestSEL108CobroCombinadoYSnapshots(t *testing.T) {
	f := newSprint3Fixture(t)
	unitPrice, poundPrice, meterPrice := 10.0, 20.0, 30.0
	_, err := f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=$2, modalidad_cobro='UNIDAD' WHERE id=$1`, f.typeID, unitPrice)
	mustSprint3(t, err)

	typeRepo := NewTipoPrendaRepository(f.db)
	canasta, err := typeRepo.Create(f.ctx, &models.TipoPrendaCreateRequest{Nombre: "Canasta", Precio: &poundPrice, ModalidadCobro: models.ModalidadCobroLibra})
	mustSprint3(t, err)
	alfombra, err := typeRepo.Create(f.ctx, &models.TipoPrendaCreateRequest{Nombre: "Alfombra", Precio: &meterPrice, ModalidadCobro: models.ModalidadCobroMetro})
	mustSprint3(t, err)

	request := &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{
		{TipoPrendaID: f.typeID, Cantidad: 2, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}}},
		{TipoPrendaID: canasta.ID, Cantidad: 2.5, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}}},
		{TipoPrendaID: alfombra.ID, Cantidad: 1.25, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[1]}}},
	}}
	order, err := NewPedidoRepository(f.db).Create(f.ctx, request, f.userID)
	mustSprint3(t, err)
	if order.Total != 107.5 {
		t.Fatalf("total combinado=%v", order.Total)
	}
	wantModes := []models.ModalidadCobro{models.ModalidadCobroUnidad, models.ModalidadCobroLibra, models.ModalidadCobroMetro}
	for i, garment := range order.Prendas {
		if garment.ModalidadCobroAplicada != wantModes[i] || garment.PrecioAplicado == nil {
			t.Fatalf("snapshot prenda %d: %+v", i, garment)
		}
	}

	changed := 99.0
	newMode := models.ModalidadCobroMetro
	_, err = typeRepo.Update(f.ctx, canasta.ID, &models.TipoPrendaUpdateRequest{Precio: &changed, ModalidadCobro: &newMode})
	mustSprint3(t, err)
	detail, err := NewPedidoRepository(f.db).GetDetalle(f.ctx, order.ID)
	mustSprint3(t, err)
	if detail.Total != 107.5 || detail.Prendas[1].ModalidadCobroAplicada != models.ModalidadCobroLibra || *detail.Prendas[1].PrecioAplicado != poundPrice {
		t.Fatalf("se altero el historico: %+v", detail.Prendas[1])
	}
}

func TestSEL108UnidadRechazaDecimalesYMedidasLosAceptan(t *testing.T) {
	f := newSprint3Fixture(t)
	_, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 1.5}}}, f.userID)
	if !errors.Is(err, ErrMedidaInvalidaModalidad) {
		t.Fatalf("UNIDAD decimal: %v", err)
	}

	price := 12.0
	tipo, err := NewTipoPrendaRepository(f.db).Create(f.ctx, &models.TipoPrendaCreateRequest{Nombre: "Peso", Precio: &price, ModalidadCobro: models.ModalidadCobroLibra})
	mustSprint3(t, err)
	order, err := NewPedidoRepository(f.db).Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{TipoPrendaID: tipo.ID, Cantidad: 0.125}}}, f.userID)
	mustSprint3(t, err)
	if order.Total != 1.5 || order.Prendas[0].Cantidad != 0.125 {
		t.Fatalf("cobro decimal: %+v", order)
	}
	added, err := NewPrendaRepository(f.db).CreateMany(f.ctx, order.ID, []models.PrendaCreateRequest{{TipoPrendaID: tipo.ID, Cantidad: 0.375}})
	mustSprint3(t, err)
	if len(added) != 1 || added[0].ModalidadCobroAplicada != models.ModalidadCobroLibra || added[0].Cantidad != 0.375 {
		t.Fatalf("alta sobre pedido existente: %+v", added)
	}
	f.total(t, order.ID, 6)

	if _, err = f.db.Exec(f.ctx, `INSERT INTO prendas(pedido_id,tipo_prenda_id,cantidad,precio_aplicado,modalidad_cobro_aplicada) VALUES($1,$2,1.0001,12,'LIBRA')`, order.ID, tipo.ID); err == nil {
		t.Fatal("la base acepto mas de tres decimales")
	}
}
