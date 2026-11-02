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

// Varios clientes sin correo deben poder coexistir (email es UNIQUE).
func TestClienteSinEmailPermiteVarios(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewClienteRepository(f.db)
	var antes int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM clientes WHERE email IS NULL`).Scan(&antes))
	for _, apellido := range []string{"Uno", "Dos", "Tres"} {
		c, err := repo.Create(f.ctx, &models.ClienteCreateRequest{Nombre: "Sin", Apellido: apellido, Telefono: "12345678", Email: "  ", Direccion: ""})
		mustSprint3(t, err)
		if c.Email != "" || c.Direccion != "" {
			t.Fatalf("cliente inesperado: %+v", c)
		}
	}
	conCorreo, err := repo.Create(f.ctx, &models.ClienteCreateRequest{Nombre: "Con", Apellido: "Correo", Telefono: "12345678", Email: "con@correo.test"})
	mustSprint3(t, err)
	vacio := ""
	actualizado, err := repo.Update(f.ctx, conCorreo.ID, &models.ClienteUpdateRequest{Email: &vacio, Direccion: &vacio})
	mustSprint3(t, err)
	if actualizado.Email != "" {
		t.Fatalf("el email vacío debe quitar el correo: %+v", actualizado)
	}
	var nulos int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM clientes WHERE email IS NULL`).Scan(&nulos))
	if nulos != antes+4 {
		t.Fatalf("clientes sin correo guardados como NULL: %d", nulos)
	}
}
