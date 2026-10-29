package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
)

func TestUsuarioRepositoryGestionaRolesYClienteRelacionado(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewUsuarioRepository(f.db)
	hash, err := auth.HashPassword("secreto-seguro")
	mustSprint3(t, err)
	cliente, err := repo.Create(f.ctx, &models.UsuarioCreateRequest{RolID: 4, Nombre: "Ana", Apellido: "Lopez", Email: "ana@selava.local", Telefono: "55551234", Direccion: "Zona 1"}, hash)
	mustSprint3(t, err)
	if cliente.ClienteID == nil || cliente.Rol != "Cliente" || cliente.PasswordHash == "secreto-seguro" {
		t.Fatalf("cliente unificado incorrecto: %#v", cliente)
	}
	if _, err = repo.Create(f.ctx, &models.UsuarioCreateRequest{RolID: 1, Nombre: "Otro", Apellido: "Admin", Email: "ANA@SELAVA.LOCAL", Telefono: ""}, hash); !errors.Is(err, ErrEmailUsuarioEnUso) {
		t.Fatalf("duplicado global: %v", err)
	}
	activo := false
	actualizado, err := repo.Update(f.ctx, cliente.ID, &models.UsuarioUpdateRequest{Activo: &activo})
	mustSprint3(t, err)
	if actualizado.Activo {
		t.Fatal("la cuenta no fue desactivada")
	}
	usuarios, err := repo.List(f.ctx)
	mustSprint3(t, err)
	if len(usuarios) < 2 {
		t.Fatalf("listado incompleto: %d", len(usuarios))
	}
}

func TestPasswordResetTokenEsDeUnSoloUso(t *testing.T) {
	f := newSprint3Fixture(t)
	repo := NewPasswordResetRepository(f.db)
	const tokenHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mustSprint3(t, repo.Crear(f.ctx, f.userID, tokenHash, time.Now().Add(30*time.Minute)))
	mustSprint3(t, repo.Validar(f.ctx, tokenHash))
	mustSprint3(t, repo.Consumir(f.ctx, tokenHash, "nuevo-hash-seguro"))
	if err := repo.Validar(f.ctx, tokenHash); !errors.Is(err, ErrTokenRecuperacionInvalido) {
		t.Fatalf("token reutilizable: %v", err)
	}
	var hash string
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT password_hash FROM usuarios WHERE id=$1`, f.userID).Scan(&hash))
	if hash != "nuevo-hash-seguro" {
		t.Fatalf("password no actualizado: %s", hash)
	}
}
