package repository

import (
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

// Clientes con email o dirección NULL (permitido por el esquema) no deben romper
// el listado ni la consulta individual del panel.
func TestClienteListadoToleraEmailYDireccionNulos(t *testing.T) {
	f := newSprint3Fixture(t)
	var id int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO clientes(nombre,apellido,telefono,email,direccion) VALUES('Sin','Datos','12345678',NULL,NULL) RETURNING id`).Scan(&id))
	repo := NewClienteRepository(f.db)
	clientes, err := repo.List(f.ctx)
	mustSprint3(t, err)
	found := false
	for _, c := range clientes {
		if c.ID == id {
			found = c.Email == "" && c.Direccion == ""
		}
	}
	if !found || len(clientes) < 2 {
		t.Fatalf("cliente con nulos no listado correctamente: %+v", clientes)
	}
	c, err := repo.GetByID(f.ctx, id)
	mustSprint3(t, err)
	if c.Email != "" || c.Direccion != "" {
		t.Fatalf("cliente inesperado: %+v", c)
	}
	nombre := "Con"
	actualizado, err := repo.Update(f.ctx, id, &models.ClienteUpdateRequest{Nombre: &nombre})
	mustSprint3(t, err)
	if actualizado.Nombre != "Con" {
		t.Fatalf("actualización inesperada: %+v", actualizado)
	}
}
