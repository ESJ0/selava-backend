package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestPedidosClienteAislanListadoDetalleEHistorial(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPedidoRepository(f.db)
	propio := f.order(t, 1)

	var otroClienteID int
	mustSprint3(t, f.db.QueryRow(f.ctx,
		`INSERT INTO clientes(nombre, apellido, telefono) VALUES('Otro','Cliente','11111111') RETURNING id`,
	).Scan(&otroClienteID))
	otro, err := repo.Create(f.ctx, &models.PedidoCreateRequest{
		ClienteID: otroClienteID,
		Prendas:   []models.PrendaCreateRequest{{TipoPrendaID: f.typeID, Cantidad: 1}},
	}, f.userID)
	mustSprint3(t, err)

	listado, err := repo.List(f.ctx, models.PedidoListFilter{ClienteID: f.clientID, Pagina: 1, Limite: 20, Orden: "recientes"})
	mustSprint3(t, err)
	if listado.Total != 1 || len(listado.Pedidos) != 1 || listado.Pedidos[0].ID != propio.ID {
		t.Fatalf("listado propio inesperado: %+v (pedido ajeno %d)", listado, otro.ID)
	}

	if _, err = repo.GetDetalleCliente(f.ctx, propio.ID, f.clientID); err != nil {
		t.Fatalf("detalle propio: %v", err)
	}
	if _, err = repo.GetDetalleCliente(f.ctx, otro.ID, f.clientID); !errors.Is(err, ErrPedidoNoEncontrado) {
		t.Fatalf("detalle ajeno debe ocultarse como no encontrado: %v", err)
	}
	if _, err = repo.GetHistorialEstadosCliente(f.ctx, propio.ID, f.clientID); err != nil {
		t.Fatalf("historial propio: %v", err)
	}
	if _, err = repo.GetHistorialEstadosCliente(f.ctx, otro.ID, f.clientID); !errors.Is(err, ErrPedidoNoEncontrado) {
		t.Fatalf("historial ajeno debe ocultarse como no encontrado: %v", err)
	}
}
