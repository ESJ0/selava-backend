package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
)

type fakeCuentaClienteRepository struct {
	registro     *models.ClienteRegistrado
	err          error
	registroReq  *models.ClienteRegistroRequest
	passwordHash string
	ultimoAcceso int
	ultimoAccErr error
}

func (f *fakeCuentaClienteRepository) Registrar(_ context.Context, req *models.ClienteRegistroRequest, passwordHash string) (*models.ClienteRegistrado, error) {
	copia := *req
	f.registroReq = &copia
	f.passwordHash = passwordHash
	return f.registro, f.err
}

func (f *fakeCuentaClienteRepository) GetByEmail(_ context.Context, email string) (*models.ClienteRegistrado, error) {
	f.registroReq = &models.ClienteRegistroRequest{Email: email}
	return f.registro, f.err
}

func (f *fakeCuentaClienteRepository) MarcarUltimoAcceso(_ context.Context, cuentaID int) error {
	f.ultimoAcceso = cuentaID
	return f.ultimoAccErr
}

func TestClienteAuthServiceRegistrarNormalizaYHasheaPassword(t *testing.T) {
	repo := &fakeCuentaClienteRepository{registro: &models.ClienteRegistrado{}}
	servicio := NewClienteAuthService(repo, "secret")
	req := &models.ClienteRegistroRequest{
		Nombre: " Ana ", Apellido: " Lopez ", Telefono: " 5555-1234 ",
		Email: " ANA@EXAMPLE.COM ", Direccion: " Zona 1 ", Password: "secreto-seguro",
	}

	if _, err := servicio.Registrar(context.Background(), req); err != nil {
		t.Fatalf("Registrar() error = %v", err)
	}
	if repo.registroReq.Email != "ana@example.com" || repo.registroReq.Nombre != "Ana" || repo.registroReq.Direccion != "Zona 1" {
		t.Fatalf("registro no normalizado: %#v", repo.registroReq)
	}
	if repo.passwordHash == req.Password || !auth.CheckPassword(repo.passwordHash, req.Password) {
		t.Fatal("el password no fue protegido con bcrypt")
	}
}

func TestClienteAuthServiceLoginGeneraTokenDeCliente(t *testing.T) {
	hash, err := auth.HashPassword("secreto-seguro")
	if err != nil {
		t.Fatal(err)
	}
	repo := &fakeCuentaClienteRepository{registro: &models.ClienteRegistrado{
		Cliente: models.Cliente{ID: 42, Activo: true},
		Cuenta:  models.CuentaCliente{ID: 9, ClienteID: 42, PasswordHash: hash, Activo: true},
	}}
	servicio := NewClienteAuthService(repo, "secret")

	token, err := servicio.Login(context.Background(), &models.ClienteLoginRequest{Email: " CLIENTE@EXAMPLE.COM ", Password: "secreto-seguro"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	claims, err := auth.ValidateClienteToken("secret", token)
	if err != nil || claims.ClienteID != 42 {
		t.Fatalf("token de cliente invalido: claims=%#v err=%v", claims, err)
	}
	if repo.registroReq.Email != "cliente@example.com" || repo.ultimoAcceso != 9 {
		t.Fatalf("login no actualizo auditoria: repo=%#v", repo)
	}
}

func TestClienteAuthServiceLoginNoRevelaCuentaInexistenteOInactiva(t *testing.T) {
	hash, err := auth.HashPassword("password-correcto")
	if err != nil {
		t.Fatal(err)
	}
	casos := []struct {
		nombre string
		repo   *fakeCuentaClienteRepository
	}{
		{"inexistente", &fakeCuentaClienteRepository{err: repository.ErrCuentaClienteNoEncontrada}},
		{"cuenta inactiva", &fakeCuentaClienteRepository{registro: &models.ClienteRegistrado{Cliente: models.Cliente{Activo: true}, Cuenta: models.CuentaCliente{Activo: false}}}},
		{"cliente inactivo", &fakeCuentaClienteRepository{registro: &models.ClienteRegistrado{Cliente: models.Cliente{Activo: false}, Cuenta: models.CuentaCliente{Activo: true}}}},
		{"password incorrecto", &fakeCuentaClienteRepository{registro: &models.ClienteRegistrado{Cliente: models.Cliente{Activo: true}, Cuenta: models.CuentaCliente{Activo: true, PasswordHash: hash}}}},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := NewClienteAuthService(caso.repo, "secret").Login(context.Background(), &models.ClienteLoginRequest{Email: "cliente@example.com", Password: "secreto-seguro"})
			if !errors.Is(err, ErrCredencialesClienteInvalidas) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
