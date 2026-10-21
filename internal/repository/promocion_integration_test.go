package repository

import (
	"errors"
	"github.com/ESJ0/selava-backend/internal/models"
	"os"
	"testing"
	"time"
)

func promoRequest(f *sprint3Fixture, mode models.ModalidadCobro, required, price float64) models.PromocionRequest {
	return models.PromocionRequest{Nombre: "Paquete historico", Descripcion: "Dos servicios sin recargo", TipoPrendaID: f.typeID, ModalidadCobro: mode, MedidaRequerida: required, PrecioPromocional: price, FechaInicio: time.Now().Add(-time.Hour), FechaFin: time.Now().Add(time.Hour), Activo: true}
}
func TestSEL109CRUD(t *testing.T) {
	f := newSprint3Fixture(t)
	r := NewPromocionRepository(f.db)
	req := promoRequest(f, models.ModalidadCobroUnidad, 2, 40)
	list, err := r.List(f.ctx)
	mustSprint3(t, err)
	if len(list) != 0 {
		t.Fatal(list)
	}
	p, err := r.Save(f.ctx, 0, &req)
	mustSprint3(t, err)
	got, err := r.GetByID(f.ctx, p.ID)
	mustSprint3(t, err)
	if got.Nombre != req.Nombre || got.TipoPrendaNombre != "Camisa" {
		t.Fatal(got)
	}
	req.Nombre = "Editada"
	req.PrecioPromocional = 35
	got, err = r.Save(f.ctx, p.ID, &req)
	mustSprint3(t, err)
	if got.Nombre != "Editada" || got.PrecioPromocional != 35 {
		t.Fatal(got)
	}
	for _, active := range []bool{false, true} {
		got, err = r.SetActivo(f.ctx, p.ID, active)
		mustSprint3(t, err)
		if got.Activo != active {
			t.Fatal(got)
		}
	}
	list, err = r.List(f.ctx)
	mustSprint3(t, err)
	if len(list) != 1 {
		t.Fatal(list)
	}
	req.TipoPrendaID = 999999
	if _, err = r.Save(f.ctx, 0, &req); !errors.Is(err, ErrTipoPrendaNoEncontrado) {
		t.Fatal(err)
	}
	req.TipoPrendaID = f.typeID
	req.ModalidadCobro = models.ModalidadCobroLibra
	if _, err = r.Save(f.ctx, 0, &req); !errors.Is(err, ErrPromocionInvalida) {
		t.Fatal(err)
	}
	if _, err = r.GetByID(f.ctx, 999999); !errors.Is(err, ErrPromocionNoEncontrada) {
		t.Fatal(err)
	}
}
func TestSEL109AgrupaUnidadesYSnapshot(t *testing.T) {
	f := newSprint3Fixture(t)
	_, err := f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=25 WHERE id=$1`, f.typeID)
	mustSprint3(t, err)
	pr := NewPromocionRepository(f.db)
	req := promoRequest(f, models.ModalidadCobroUnidad, 2, 40)
	p, err := pr.Save(f.ctx, 0, &req)
	mustSprint3(t, err)
	orderReq := &models.PedidoCreateRequest{ClienteID: f.clientID, Promociones: []models.PromocionSeleccion{{TipoPrendaID: f.typeID, PromocionID: p.ID}}}
	for i := 0; i < 5; i++ {
		orderReq.Prendas = append(orderReq.Prendas, models.PrendaCreateRequest{TipoPrendaID: f.typeID, Cantidad: 1, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}, {ServicioID: f.services[1]}}})
	}
	quote, err := pr.Cotizar(f.ctx, orderReq)
	mustSprint3(t, err)
	if quote.Total != 105 || quote.Grupos[0].Aplicada.Aplicaciones != 2 {
		t.Fatal(quote)
	}
	order, err := NewPedidoRepository(f.db).Create(f.ctx, orderReq, f.userID)
	mustSprint3(t, err)
	f.total(t, order.ID, 105)
	if len(order.PromocionesAplicadas) != 1 || len(order.PromocionesAplicadas[0].PrendaIDs) != 5 {
		t.Fatal(order)
	}
	req.Nombre = "Cambio posterior"
	req.PrecioPromocional = 1
	_, err = pr.Save(f.ctx, p.ID, &req)
	mustSprint3(t, err)
	_, err = pr.SetActivo(f.ctx, p.ID, false)
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=999,modalidad_cobro='LIBRA' WHERE id=$1`, f.typeID)
	mustSprint3(t, err)
	detail, err := NewPedidoRepository(f.db).GetDetalleCliente(f.ctx, order.ID, f.clientID)
	mustSprint3(t, err)
	a := detail.PromocionesAplicadas[0]
	if detail.Total != 105 || a.Nombre != "Paquete historico" || a.PrecioPromocional != 40 || a.ModalidadCobroAplicada != models.ModalidadCobroUnidad || a.SubtotalPromocional != 80 || a.SubtotalNormal != 25 || a.Sobrante != 1 {
		t.Fatalf("historico alterado %+v", a)
	}
	// Recalcular por cambios de servicios y agregar prendas no vuelve a aplicar
	// promociones ni pierde el ajuste congelado del pedido original.
	r := NewPrendaRepository(f.db)
	_, err = r.AddServicio(f.ctx, order.Prendas[0].ID, f.services[2])
	mustSprint3(t, err)
	mustSprint3(t, r.RemoveServicio(f.ctx, order.Prendas[0].ID, f.services[0]))
	f.total(t, order.ID, 105)
	_, err = r.CreateMany(f.ctx, order.ID, []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 1}})
	mustSprint3(t, err)
	f.total(t, order.ID, 1104)
	down, err := os.ReadFile("../../migrations/025_promociones.down.sql")
	mustSprint3(t, err)
	if _, err = f.db.Exec(f.ctx, string(down)); err == nil {
		t.Fatal("rollback borro snapshots")
	}
}
func TestSEL109LibraMetroYSeleccionInvalida(t *testing.T) {
	f := newSprint3Fixture(t)
	r := NewPromocionRepository(f.db)
	for _, tc := range []struct {
		mode                                    models.ModalidadCobro
		required, price, measure, normal, total float64
	}{{models.ModalidadCobroLibra, 5, 50, 12.5, 12, 130}, {models.ModalidadCobroMetro, 2, 60, 5.5, 40, 180}} {
		_, err := f.db.Exec(f.ctx, `UPDATE tipos_prenda SET precio=$2,modalidad_cobro=$3 WHERE id=$1`, f.typeID, tc.normal, tc.mode)
		mustSprint3(t, err)
		req := promoRequest(f, tc.mode, tc.required, tc.price)
		p, err := r.Save(f.ctx, 0, &req)
		mustSprint3(t, err)
		orderReq := &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: tc.measure, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}}}}, Promociones: []models.PromocionSeleccion{{TipoPrendaID: f.typeID, PromocionID: p.ID}}}
		order, err := NewPedidoRepository(f.db).Create(f.ctx, orderReq, f.userID)
		mustSprint3(t, err)
		f.total(t, order.ID, tc.total)
		if order.PromocionesAplicadas[0].Aplicaciones != 2 {
			t.Fatal(order)
		}
		orderReq.Promociones = append(orderReq.Promociones, orderReq.Promociones[0])
		if _, err = r.Cotizar(f.ctx, orderReq); !errors.Is(err, ErrPromocionInvalida) {
			t.Fatal(err)
		}
		orderReq.Promociones = orderReq.Promociones[:1]
		for _, state := range []string{"inactiva", "futura", "vencida", "otra_modalidad", "insuficiente"} {
			req = promoRequest(f, tc.mode, tc.required, tc.price)
			switch state {
			case "inactiva":
				req.Activo = false
			case "futura":
				req.FechaInicio = time.Now().Add(time.Minute)
			case "vencida":
				req.FechaFin = time.Now().Add(-time.Minute)
			case "otra_modalidad":
				_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET modalidad_cobro='UNIDAD' WHERE id=$1`, f.typeID)
				mustSprint3(t, err)
			case "insuficiente":
				orderReq.Prendas[0].Cantidad = 0.001
			}
			if state != "otra_modalidad" {
				_, err = r.Save(f.ctx, p.ID, &req)
				mustSprint3(t, err)
			}
			if _, err = r.Cotizar(f.ctx, orderReq); err == nil {
				t.Fatal("seleccion invalida", state)
			}
			_, err = f.db.Exec(f.ctx, `UPDATE tipos_prenda SET modalidad_cobro=$2 WHERE id=$1`, f.typeID, tc.mode)
			mustSprint3(t, err)
		}
	}
}
