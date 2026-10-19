package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/service"
)

const clienteSecuritySecret = "cliente-security-test-secret"

// memoriaCuentaClienteRepository reemplaza solo PostgreSQL. El resto del flujo
// HTTP usa controlador, servicio, bcrypt, JWT, middleware y rutas reales.
type memoriaCuentaClienteRepository struct {
	cuentas      map[string]*models.ClienteRegistrado
	siguienteID  int
	ultimoAcceso int
}

func nuevaMemoriaCuentaClienteRepository() *memoriaCuentaClienteRepository {
	return &memoriaCuentaClienteRepository{cuentas: make(map[string]*models.ClienteRegistrado), siguienteID: 1}
}

func (r *memoriaCuentaClienteRepository) Registrar(_ context.Context, req *models.ClienteRegistroRequest, passwordHash string) (*models.ClienteRegistrado, error) {
	if _, existe := r.cuentas[req.Email]; existe {
		return nil, repository.ErrEmailCuentaClienteEnUso
	}
	id := r.siguienteID
	r.siguienteID++
	ahora := time.Now()
	registro := &models.ClienteRegistrado{
		Cliente: models.Cliente{
			ID: id, Nombre: req.Nombre, Apellido: req.Apellido,
			Telefono: req.Telefono, Email: req.Email, Direccion: req.Direccion,
			Activo: true, CreatedAt: ahora, UpdatedAt: ahora,
		},
		Cuenta: models.CuentaCliente{
			ID: id, ClienteID: id, Email: req.Email, PasswordHash: passwordHash,
			Activo: true, CreatedAt: ahora, UpdatedAt: ahora,
		},
	}
	r.cuentas[req.Email] = registro
	return registro, nil
}

func (r *memoriaCuentaClienteRepository) GetByEmail(_ context.Context, email string) (*models.ClienteRegistrado, error) {
	registro, existe := r.cuentas[email]
	if !existe {
		return nil, repository.ErrCuentaClienteNoEncontrada
	}
	return registro, nil
}

func (r *memoriaCuentaClienteRepository) MarcarUltimoAcceso(_ context.Context, cuentaID int) error {
	r.ultimoAcceso = cuentaID
	return nil
}

func nuevoRouterClienteEnMemoria() (http.Handler, *memoriaCuentaClienteRepository) {
	repo := nuevaMemoriaCuentaClienteRepository()
	authService := service.NewClienteAuthService(repo, clienteSecuritySecret)
	authController := controller.NewClienteAuthController(authService)
	router := NewRouterWithClienteAuth(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, authController, clienteSecuritySecret, "")
	return router, repo
}

func ejecutarJSON(router http.Handler, method, path, body string, token ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func registrarClienteHTTP(t *testing.T, router http.Handler, email string) int {
	t.Helper()
	body := `{"nombre":"Ana","apellido":"Lopez","telefono":"5555-1234","email":"` + email + `","direccion":"Zona 1","password":"secreto-seguro"}`
	res := ejecutarJSON(router, http.MethodPost, "/api/auth/clientes/registro", body)
	if res.Code != http.StatusCreated {
		t.Fatalf("registro status=%d body=%s", res.Code, res.Body.String())
	}
	var respuesta models.ClienteRegistroResponse
	if err := json.Unmarshal(res.Body.Bytes(), &respuesta); err != nil {
		t.Fatal(err)
	}
	if respuesta.Cliente.ID <= 0 || respuesta.Cuenta.ClienteID != respuesta.Cliente.ID {
		t.Fatalf("respuesta de registro invalida: %#v", respuesta)
	}
	if strings.Contains(res.Body.String(), "password") {
		t.Fatalf("el registro expuso credenciales: %s", res.Body.String())
	}
	return respuesta.Cliente.ID
}

func loginClienteHTTP(t *testing.T, router http.Handler, email, password string) string {
	t.Helper()
	body := `{"email":"` + email + `","password":"` + password + `"}`
	res := ejecutarJSON(router, http.MethodPost, "/api/auth/clientes/login", body)
	if res.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", res.Code, res.Body.String())
	}
	var respuesta struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &respuesta); err != nil {
		t.Fatal(err)
	}
	if respuesta.Token == "" {
		t.Fatal("login no devolvio token")
	}
	return respuesta.Token
}

func TestFlujoRegistroYLoginCliente(t *testing.T) {
	router, repo := nuevoRouterClienteEnMemoria()
	clienteID := registrarClienteHTTP(t, router, "ana@example.com")

	duplicado := ejecutarJSON(router, http.MethodPost, "/api/auth/clientes/registro", `{"nombre":"Otra","apellido":"Persona","telefono":"5555-9876","email":"ANA@EXAMPLE.COM","password":"otro-secreto"}`)
	if duplicado.Code != http.StatusConflict {
		t.Fatalf("duplicado status=%d body=%s", duplicado.Code, duplicado.Body.String())
	}

	invalido := ejecutarJSON(router, http.MethodPost, "/api/auth/clientes/login", `{"email":"ana@example.com","password":"incorrecta"}`)
	if invalido.Code != http.StatusUnauthorized {
		t.Fatalf("credenciales invalidas status=%d", invalido.Code)
	}

	registroInvalido := ejecutarJSON(router, http.MethodPost, "/api/auth/clientes/registro", `{"email":"correo-invalido","password":"corta"}`)
	if registroInvalido.Code != http.StatusUnprocessableEntity {
		t.Fatalf("registro invalido status=%d body=%s", registroInvalido.Code, registroInvalido.Body.String())
	}

	token := loginClienteHTTP(t, router, "ANA@EXAMPLE.COM", "secreto-seguro")
	claims, err := auth.ValidateClienteToken(clienteSecuritySecret, token)
	if err != nil || claims.ClienteID != clienteID || claims.Actor != auth.ActorCliente {
		t.Fatalf("claims invalidos: %#v err=%v", claims, err)
	}
	if repo.ultimoAcceso != clienteID {
		t.Fatalf("ultimo acceso=%d, esperado=%d", repo.ultimoAcceso, clienteID)
	}

	protegido := ejecutarJSON(router, http.MethodGet, "/api/cliente/me", "", token)
	if protegido.Code != http.StatusOK || !strings.Contains(protegido.Body.String(), `"id":`+strconv.Itoa(clienteID)) {
		t.Fatalf("acceso protegido status=%d body=%s", protegido.Code, protegido.Body.String())
	}
}
