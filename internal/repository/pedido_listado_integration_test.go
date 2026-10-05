package repository

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/validator"
)

func TestPedidoListadoFiltersAndPagination(t *testing.T) {
	f := newSprint3Fixture(t)
	_, err := f.db.Exec(f.ctx, "UPDATE clientes SET nombre='Ana', apellido='Pérez', telefono='55123456' WHERE id=$1", f.clientID)
	mustSprint3(t, err)
	var secondClient int
	mustSprint3(t, f.db.QueryRow(f.ctx, "INSERT INTO clientes(nombre,apellido,telefono) VALUES('Luis','Gómez_50%','88770000') RETURNING id").Scan(&secondClient))
	dates := []string{"2026-10-04T23:59:59-06:00", "2026-10-05T00:00:00-06:00", "2026-10-05T23:59:59.999999-06:00", "2026-10-06T00:00:00-06:00"}
	ids := []int{}
	for i, value := range dates {
		order := f.order(t, 1)
		ids = append(ids, order.ID)
		date, err := time.Parse(time.RFC3339Nano, value)
		mustSprint3(t, err)
		client, state, active := f.clientID, f.state(t, "Recibido"), true
		if i == 2 {
			client, state, active = secondClient, f.state(t, "Cancelado"), false
		}
		_, err = f.db.Exec(f.ctx, "UPDATE pedidos SET fecha_recibido=$1,cliente_id=$2,estado_actual_id=$3,activo=$4 WHERE id=$5", date, client, state, active, order.ID)
		mustSprint3(t, err)
	}
	repo := NewPedidoRepository(f.db)
	for _, tc := range []struct {
		name  string
		req   models.PedidoListRequest
		ids   []int
		total int
	}{
		{"all includes cancelled", models.PedidoListRequest{}, []int{ids[3], ids[2], ids[1], ids[0]}, 4},
		{"first page", models.PedidoListRequest{Limite: "2"}, []int{ids[3], ids[2]}, 4},
		{"second page", models.PedidoListRequest{Limite: "2", Pagina: "2"}, []int{ids[1], ids[0]}, 4},
		{"oldest", models.PedidoListRequest{Limite: "2", Orden: "antiguos"}, []int{ids[0], ids[1]}, 4},
		{"inclusive local day", models.PedidoListRequest{FechaDesde: "2026-10-05", FechaHasta: "2026-10-05"}, []int{ids[2], ids[1]}, 2},
		{"from only", models.PedidoListRequest{FechaDesde: "2026-10-05"}, []int{ids[3], ids[2], ids[1]}, 3},
		{"until only", models.PedidoListRequest{FechaHasta: "2026-10-05"}, []int{ids[2], ids[1], ids[0]}, 3},
		{"name and date", models.PedidoListRequest{Buscar: "ana p", FechaDesde: "2026-10-05", FechaHasta: "2026-10-05"}, []int{ids[1]}, 1},
		{"phone", models.PedidoListRequest{Buscar: "8877"}, []int{ids[2]}, 1},
		{"receipt number", models.PedidoListRequest{Buscar: fmt.Sprintf("#SLV-%04d", ids[1])}, []int{ids[1]}, 1},
		{"cancelled status", models.PedidoListRequest{EstadoID: fmt.Sprint(f.state(t, "Cancelado"))}, []int{ids[2]}, 1},
		{"literal percent", models.PedidoListRequest{Buscar: "%"}, []int{ids[2]}, 1},
		{"literal underscore", models.PedidoListRequest{Buscar: "_"}, []int{ids[2]}, 1},
		{"sql injection", models.PedidoListRequest{Buscar: "' OR 1=1--"}, []int{}, 0},
		{"large number", models.PedidoListRequest{Buscar: "9999999999999999"}, []int{}, 0},
		{"out of range page", models.PedidoListRequest{Pagina: "5", Limite: "2"}, []int{}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filter, errs := validator.ParsePedidoList(tc.req)
			if errs.HasErrors() {
				t.Fatal(errs)
			}
			result, err := repo.List(f.ctx, filter)
			mustSprint3(t, err)
			got := []int{}
			for _, item := range result.Pedidos {
				got = append(got, item.ID)
				if item.Cliente.ID <= 0 || item.EstadoActual.ID <= 0 || item.Total != 10.25 {
					t.Fatalf("resumen incompleto: %+v", item)
				}
			}
			if !reflect.DeepEqual(got, tc.ids) || result.Total != tc.total || result.Pedidos == nil || result.Pagina != filter.Pagina || result.Limite != filter.Limite {
				t.Fatalf("ids=%v expected=%v response=%+v", got, tc.ids, result)
			}
		})
	}
}

func TestPedidoListadoEmptyAndStableOrder(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPedidoRepository(f.db)
	filter, _ := validator.ParsePedidoList(models.PedidoListRequest{Limite: "1"})
	empty, err := repo.List(f.ctx, filter)
	mustSprint3(t, err)
	if empty.Total != 0 || empty.Pedidos == nil || len(empty.Pedidos) != 0 {
		t.Fatalf("empty=%+v", empty)
	}
	first, second := f.order(t, 1), f.order(t, 1)
	_, err = f.db.Exec(f.ctx, "UPDATE pedidos SET fecha_recibido='2026-10-05T12:00:00Z' WHERE id IN ($1,$2)", first.ID, second.ID)
	mustSprint3(t, err)
	for page, id := range []int{second.ID, first.ID} {
		filter.Pagina = page + 1
		result, err := repo.List(f.ctx, filter)
		mustSprint3(t, err)
		if result.Total != 2 || len(result.Pedidos) != 1 || result.Pedidos[0].ID != id {
			t.Fatalf("page=%d result=%+v", page+1, result)
		}
	}
}
