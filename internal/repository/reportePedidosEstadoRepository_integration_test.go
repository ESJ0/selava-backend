package repository

import "testing"

func TestReportePedidosEstadoCuentaEstadoActualEIncluyeEstadosVacios(t *testing.T) {
	f := newSprint3Fixture(t)
	pedidoRecibido := f.order(t, 1)
	pedidoRackeado := f.order(t, 1)
	rackeadoID := f.state(t, "Rackeado")
	_, err := f.db.Exec(f.ctx, `UPDATE pedidos SET estado_actual_id=$1 WHERE id=$2`, rackeadoID, pedidoRackeado.ID)
	mustSprint3(t, err)

	reporte, err := NewReportePedidosEstadoRepository(f.db).Get(f.ctx)
	mustSprint3(t, err)

	if reporte.TotalPedidos != 2 {
		t.Fatalf("total_pedidos=%d, esperado=2", reporte.TotalPedidos)
	}
	if len(reporte.PedidosPorEstado) < 3 {
		t.Fatalf("se esperaban todos los estados, se obtuvo %+v", reporte.PedidosPorEstado)
	}
	cantidades := make(map[string]int)
	for _, item := range reporte.PedidosPorEstado {
		cantidades[item.Estado.Nombre] = item.CantidadPedidos
	}
	if cantidades["Recibido"] != 1 || cantidades["Rackeado"] != 1 || cantidades["Cancelado"] != 0 {
		t.Fatalf("conteos inesperados: %+v (pedido recibido %d)", cantidades, pedidoRecibido.ID)
	}
}
