package repository

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
)

func TestPedidoListadoPortalClienteSeccionesSaldoYTipos(t *testing.T) {
	f := newSprint3Fixture(t)
	var pantalonID, otroClienteID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO tipos_prenda(nombre,precio) VALUES('Pantalón',15) RETURNING id`).Scan(&pantalonID))
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO clientes(nombre,apellido,telefono) VALUES('Otro','Cliente','11111111') RETURNING id`).Scan(&otroClienteID))
	metodo := sprint4PaymentMethod(t, f, "Efectivo portal")
	pedidoRepo := NewPedidoRepository(f.db)

	type caso struct {
		estado  string
		total   float64
		pagado  float64
		tipo    int
		cliente int
		activo  bool
	}
	casos := []caso{
		{"Recibido", 100, 40, f.typeID, f.clientID, true},      // 0 activo con saldo
		{"Recibido", 50, 50, pantalonID, f.clientID, true},     // 1 activo pagado
		{"Finalizado", 80, 30.333, f.typeID, f.clientID, true}, // 2 finalizado con saldo
		{"Finalizado", 20, 20, pantalonID, f.clientID, true},   // 3 finalizado pagado
		{"Cancelado", 30, 0, f.typeID, f.clientID, true},       // 4 cancelado (sin saldo exigible)
		{"Recibido", 70, 0, f.typeID, f.clientID, false},       // 5 eliminado lógicamente
		{"Recibido", 60, 0, pantalonID, otroClienteID, true},   // 6 de otro cliente
	}
	ids := make([]int, len(casos))
	for i, c := range casos {
		order, err := pedidoRepo.Create(f.ctx, &models.PedidoCreateRequest{ClienteID: f.clientID, Prendas: []models.PrendaCreateRequest{{
			TipoPrendaID: c.tipo, Cantidad: 2, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}},
		}}}, f.userID)
		mustSprint3(t, err)
		ids[i] = order.ID
		fecha := time.Date(2026, 10, 1+i, 12, 0, 0, 0, time.UTC)
		_, err = f.db.Exec(f.ctx, `UPDATE pedidos SET total=$1, cliente_id=$2, activo=$3, fecha_recibido=$4, estado_actual_id=(SELECT id FROM estados_pedido WHERE nombre=$5) WHERE id=$6`,
			c.total, c.cliente, c.activo, fecha, c.estado, order.ID)
		mustSprint3(t, err)
		if c.pagado > 0 {
			_, err = f.db.Exec(f.ctx, `INSERT INTO pagos(pedido_id,metodo_pago_id,usuario_id,monto) VALUES($1,$2,$3,$4)`, order.ID, metodo, f.userID, c.pagado)
			mustSprint3(t, err)
		}
	}

	list := func(t *testing.T, req models.PedidoListRequest) *models.PedidoListResponse {
		t.Helper()
		filter, errs := validator.ParsePedidoList(req)
		if errs.HasErrors() {
			t.Fatal(errs)
		}
		filter.ClienteID = f.clientID
		filter.SoloVigentes = true
		result, err := pedidoRepo.List(f.ctx, filter)
		mustSprint3(t, err)
		return result
	}
	idsOf := func(result *models.PedidoListResponse) []int {
		got := []int{}
		for _, p := range result.Pedidos {
			got = append(got, p.ID)
		}
		return got
	}

	for _, tc := range []struct {
		name string
		req  models.PedidoListRequest
		want []int
	}{
		{"todos vigentes", models.PedidoListRequest{}, []int{ids[4], ids[3], ids[2], ids[1], ids[0]}},
		{"activos", models.PedidoListRequest{Seccion: "activos"}, []int{ids[1], ids[0]}},
		{"historial", models.PedidoListRequest{Seccion: "historial"}, []int{ids[4], ids[3], ids[2]}},
		{"saldo pendiente sin cancelados", models.PedidoListRequest{Saldo: "pendiente"}, []int{ids[2], ids[0]}},
		{"saldo pendiente prioriza activos", models.PedidoListRequest{Saldo: "pendiente", Orden: "prioridad"}, []int{ids[0], ids[2]}},
		{"pagados", models.PedidoListRequest{Saldo: "pagado"}, []int{ids[3], ids[1]}},
		{"historial pagado", models.PedidoListRequest{Seccion: "historial", Saldo: "pagado"}, []int{ids[3]}},
		{"tipo de prenda", models.PedidoListRequest{TipoPrendaID: strconv.Itoa(pantalonID)}, []int{ids[3], ids[1]}},
		{"historial por tipo y fecha", models.PedidoListRequest{Seccion: "historial", TipoPrendaID: strconv.Itoa(f.typeID), FechaDesde: "2026-10-04"}, []int{ids[4]}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := idsOf(list(t, tc.req))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ids=%v esperado=%v", got, tc.want)
			}
		})
	}

	result := list(t, models.PedidoListRequest{})
	byID := map[int]models.PedidoResumen{}
	for _, p := range result.Pedidos {
		byID[p.ID] = p
	}
	check := func(id int, pagado, saldo float64, enProceso, cancelado bool) {
		t.Helper()
		p := byID[id]
		if math.Abs(p.TotalPagado-pagado) > 0.001 || math.Abs(p.SaldoPendiente-saldo) > 0.001 || p.EnProceso != enProceso || p.Cancelado != cancelado {
			t.Fatalf("pedido %d inesperado: %+v", id, p)
		}
	}
	check(ids[0], 40, 60, true, false)
	check(ids[1], 50, 0, true, false)
	check(ids[2], 30.33, 49.67, false, false)
	check(ids[3], 20, 0, false, false)
	check(ids[4], 0, 30, false, true)
	if p := byID[ids[1]]; !reflect.DeepEqual(p.TiposPrenda, []string{"Pantalón"}) || p.CantidadPrendas != 2 {
		t.Fatalf("prendas inesperadas: %+v", p)
	}

	r := result.Resumen
	if r == nil {
		t.Fatal("falta resumen")
	}
	if r.Activos != 2 || r.Historial != 3 || r.PedidosConSaldo != 2 ||
		math.Abs(r.SaldoPendiente-109.67) > 0.001 || math.Abs(r.TotalPagado-140.33) > 0.001 || math.Abs(r.TotalFacturado-250) > 0.001 {
		t.Fatalf("resumen inesperado: %+v", r)
	}
	// El resumen no depende de los filtros de la consulta.
	if filtrado := list(t, models.PedidoListRequest{Seccion: "activos"}); !reflect.DeepEqual(filtrado.Resumen, r) {
		t.Fatalf("resumen filtrado=%+v esperado=%+v", filtrado.Resumen, r)
	}

	tipos, err := pedidoRepo.TiposPrendaCliente(f.ctx, f.clientID)
	mustSprint3(t, err)
	want := []models.TipoPrendaOpcion{{ID: f.typeID, Nombre: "Camisa"}, {ID: pantalonID, Nombre: "Pantalón"}}
	if !reflect.DeepEqual(tipos, want) {
		t.Fatalf("tipos=%+v esperado=%+v", tipos, want)
	}
	tipos, err = pedidoRepo.TiposPrendaCliente(f.ctx, 999999)
	mustSprint3(t, err)
	if len(tipos) != 0 {
		t.Fatalf("cliente inexistente debe devolver lista vacía: %+v", tipos)
	}
}
