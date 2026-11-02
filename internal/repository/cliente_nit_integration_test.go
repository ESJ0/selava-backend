package repository

import (
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestClienteRepositoryGuardaActualizaYBorraNIT(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewClienteRepository(f.db)
	creado, err := repo.Create(f.ctx, &models.ClienteCreateRequest{Nombre: "Nit", Apellido: "Prueba", Telefono: "55550001", NIT: "576937-K"})
	mustSprint3(t, err)
	if creado.NIT != "576937-K" {
		t.Fatalf("NIT no guardado: %q", creado.NIT)
	}
	sinCambio, err := repo.Update(f.ctx, creado.ID, &models.ClienteUpdateRequest{Telefono: ptr("55550002")})
	mustSprint3(t, err)
	if sinCambio.NIT != "576937-K" {
		t.Fatalf("NIT perdido al actualizar otros campos: %q", sinCambio.NIT)
	}
	cf := "CF"
	actualizado, err := repo.Update(f.ctx, creado.ID, &models.ClienteUpdateRequest{NIT: &cf})
	mustSprint3(t, err)
	if actualizado.NIT != "CF" {
		t.Fatalf("NIT no actualizado: %q", actualizado.NIT)
	}
	vacio := ""
	borrado, err := repo.Update(f.ctx, creado.ID, &models.ClienteUpdateRequest{NIT: &vacio})
	mustSprint3(t, err)
	listado, err := repo.GetByID(f.ctx, creado.ID)
	mustSprint3(t, err)
	if borrado.NIT != "" || listado.NIT != "" {
		t.Fatalf("NIT no se pudo quitar: %q %q", borrado.NIT, listado.NIT)
	}
	if _, err := repo.Create(f.ctx, &models.ClienteCreateRequest{Nombre: "Mal", Apellido: "Nit", Telefono: "55550003", NIT: "12AB"}); err == nil {
		t.Fatal("la base acepto un NIT con formato invalido")
	}
}

func ptr(s string) *string { return &s }
