package repository

import (
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestCuentaClienteRepositoryRegistraPerfilYCuentaAtomicamente(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewCuentaClienteRepository(f.db)
	req := &models.ClienteRegistroRequest{
		Nombre: "Ana", Apellido: "Lopez", Telefono: "5555-1234",
		Email: "ana@example.com", Direccion: "Zona 1",
	}

	registro, err := repo.Registrar(f.ctx, req, "hash-de-prueba")
	mustSprint3(t, err)
	if registro.Cliente.ID == 0 || registro.Cuenta.ClienteID != registro.Cliente.ID {
		t.Fatalf("relacion incorrecta: %#v", registro)
	}

	obtenido, err := repo.GetByEmail(f.ctx, req.Email)
	mustSprint3(t, err)
	if obtenido.Cuenta.PasswordHash != "hash-de-prueba" || obtenido.Cliente.Email != req.Email {
		t.Fatalf("cuenta obtenida incorrecta: %#v", obtenido)
	}
	perfil, err := repo.GetClienteActivoByID(f.ctx, registro.Cliente.ID)
	mustSprint3(t, err)
	if perfil.ID != registro.Cliente.ID || perfil.Nombre != "Ana" || perfil.Email != req.Email {
		t.Fatalf("perfil de sesion incorrecto: %#v", perfil)
	}

	mustSprint3(t, repo.MarcarUltimoAcceso(f.ctx, obtenido.Cuenta.ID))
	var actualizado bool
	mustSprint3(t, f.db.QueryRow(f.ctx, `
		SELECT ultimo_acceso_at IS NOT NULL FROM usuarios WHERE id=$1
	`, obtenido.Cuenta.ID).Scan(&actualizado))
	if !actualizado {
		t.Fatal("ultimo_acceso_at no fue actualizado")
	}
}

func TestCuentaClienteRepositoryRechazaEmailDuplicado(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewCuentaClienteRepository(f.db)
	req := &models.ClienteRegistroRequest{
		Nombre: "Ana", Apellido: "Lopez", Telefono: "5555-1234", Email: "duplicado@example.com",
	}
	_, err := repo.Registrar(f.ctx, req, "hash-de-prueba")
	mustSprint3(t, err)

	req.Nombre = "Otra"
	_, err = repo.Registrar(f.ctx, req, "otro-hash")
	if !errors.Is(err, ErrEmailCuentaClienteEnUso) {
		t.Fatalf("error = %v", err)
	}

	var perfiles, cuentas int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM clientes WHERE email=$1`, req.Email).Scan(&perfiles))
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT COUNT(*) FROM usuarios WHERE email=$1 AND rol_id=4`, req.Email).Scan(&cuentas))
	if perfiles != 1 || cuentas != 1 {
		t.Fatalf("el registro duplicado dejo datos parciales: perfiles=%d cuentas=%d", perfiles, cuentas)
	}
}
