package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type fakePagoRepository struct {
	created     *models.PagoCreateRequest
	createCalls int
	err         error
	saldo       *models.SaldoPedido
	pagos       []models.PagoDetalle
	clienteID   int
}

func (r *fakePagoRepository) ListByPedido(_ context.Context, _ int) ([]models.PagoDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.pagos, nil
}

func (r *fakePagoRepository) ListByPedidoCliente(_ context.Context, _ int, clienteID int) ([]models.PagoDetalle, error) {
	r.clienteID = clienteID
	return r.ListByPedido(context.Background(), 0)
}

func (r *fakePagoRepository) GetSaldo(_ context.Context, _ int) (*models.SaldoPedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.saldo, nil
}

func (r *fakePagoRepository) GetSaldoCliente(_ context.Context, _ int, clienteID int) (*models.SaldoPedido, error) {
	r.clienteID = clienteID
	return r.GetSaldo(context.Background(), 0)
}

func (r *fakePagoRepository) Create(_ context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error) {
	r.createCalls++
	if r.err != nil {
		return nil, r.err
	}
	r.created = req
	return &models.PagoDetalle{Pago: models.Pago{ID: 1, PedidoID: pedidoID, UsuarioID: usuarioID, Monto: req.Monto, Referencia: req.Referencia}}, nil
}

func TestPagoServiceRegistrarPagoRechazaMontoNoPositivo(t *testing.T) {
	for _, test := range []struct {
		name  string
		monto float64
	}{
		{name: "cero", monto: 0},
		{name: "negativo", monto: -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakePagoRepository{}
			service := NewPagoService(repo)

			_, err := service.RegistrarPago(context.Background(), 1, &models.PagoCreateRequest{MetodoPagoID: 1, Monto: test.monto}, 1)
			var validationErrors validator.ValidationErrors
			if !errors.As(err, &validationErrors) {
				t.Fatalf("expected ValidationErrors, got %v", err)
			}
			if repo.createCalls != 0 {
				t.Fatalf("repository Create called %d times", repo.createCalls)
			}
		})
	}
}

func TestPagoServiceRegistrarPagoRechazaIdentificadoresInvalidos(t *testing.T) {
	tests := []struct {
		name       string
		pedidoID   int
		metodoID   int
		usuarioID  int
		want       error
		validation bool
	}{
		{name: "pedido", pedidoID: 0, metodoID: 1, usuarioID: 1, want: repository.ErrPedidoNoEncontrado},
		{name: "usuario", pedidoID: 1, metodoID: 1, usuarioID: 0, want: repository.ErrUsuarioNoEncontrado},
		{name: "metodo", pedidoID: 1, metodoID: 0, usuarioID: 1, validation: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakePagoRepository{}
			service := NewPagoService(repo)
			_, err := service.RegistrarPago(context.Background(), test.pedidoID, &models.PagoCreateRequest{MetodoPagoID: test.metodoID, Monto: 1}, test.usuarioID)
			if test.validation {
				var validationErrors validator.ValidationErrors
				if !errors.As(err, &validationErrors) {
					t.Fatalf("expected ValidationErrors, got %v", err)
				}
			} else if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if repo.createCalls != 0 {
				t.Fatalf("repository Create called %d times", repo.createCalls)
			}
		})
	}
}

func TestPagoServiceRegistrarPagoNormalizaReferencia(t *testing.T) {
	repo := &fakePagoRepository{}
	service := NewPagoService(repo)
	referencia := "  POS-123  "

	pago, err := service.RegistrarPago(context.Background(), 4, &models.PagoCreateRequest{MetodoPagoID: 2, Monto: 25.50, Referencia: &referencia}, 7)
	if err != nil {
		t.Fatalf("RegistrarPago() error = %v", err)
	}
	if pago.PedidoID != 4 || pago.UsuarioID != 7 {
		t.Fatalf("pago inesperado: %+v", pago)
	}
	if repo.created.Referencia == nil || *repo.created.Referencia != "POS-123" {
		t.Fatalf("referencia no normalizada: %+v", repo.created.Referencia)
	}
}

func TestPagoServiceRegistrarPagoRechazaMontoConMasDeDosDecimales(t *testing.T) {
	service := NewPagoService(&fakePagoRepository{})
	_, err := service.RegistrarPago(context.Background(), 1, &models.PagoCreateRequest{MetodoPagoID: 1, Monto: 10.125}, 1)
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("se esperaba ValidationErrors, se obtuvo %v", err)
	}
}

func TestPagoServiceRegistrarPagoPropagaConflictoDeSaldo(t *testing.T) {
	service := NewPagoService(&fakePagoRepository{err: repository.ErrPagoExcedeSaldo})
	_, err := service.RegistrarPago(context.Background(), 1, &models.PagoCreateRequest{MetodoPagoID: 1, Monto: 10}, 1)
	if !errors.Is(err, repository.ErrPagoExcedeSaldo) {
		t.Fatalf("error = %v, se esperaba ErrPagoExcedeSaldo", err)
	}
}

func TestPagoServiceObtenerSaldo(t *testing.T) {
	esperado := &models.SaldoPedido{PedidoID: 3, Total: 100, TotalPagado: 35, SaldoPendiente: 65}
	service := NewPagoService(&fakePagoRepository{saldo: esperado})
	saldo, err := service.ObtenerSaldo(context.Background(), 3)
	if err != nil {
		t.Fatalf("ObtenerSaldo() error = %v", err)
	}
	if saldo != esperado {
		t.Fatalf("saldo = %+v, se esperaba %+v", saldo, esperado)
	}
}

func TestPagoServiceObtenerSaldoRechazaIDInvalido(t *testing.T) {
	service := NewPagoService(&fakePagoRepository{})
	_, err := service.ObtenerSaldo(context.Background(), 0)
	if !errors.Is(err, repository.ErrPedidoNoEncontrado) {
		t.Fatalf("error = %v, se esperaba ErrPedidoNoEncontrado", err)
	}
}

func TestPagoServiceObtenerHistorial(t *testing.T) {
	esperados := []models.PagoDetalle{{Pago: models.Pago{ID: 1, PedidoID: 3, Monto: 20}}, {Pago: models.Pago{ID: 2, PedidoID: 3, Monto: 15}}}
	service := NewPagoService(&fakePagoRepository{pagos: esperados})
	pagos, err := service.ObtenerHistorial(context.Background(), 3)
	if err != nil {
		t.Fatalf("ObtenerHistorial() error = %v", err)
	}
	if len(pagos) != 2 || pagos[1].ID != 2 {
		t.Fatalf("historial inesperado: %+v", pagos)
	}
}

func TestPagoServiceConsultasClienteUsanIdentidadYOcultanUsuario(t *testing.T) {
	referencia := "TRX-9"
	repo := &fakePagoRepository{
		saldo: &models.SaldoPedido{PedidoID: 3, Total: 100, TotalPagado: 35, SaldoPendiente: 65},
		pagos: []models.PagoDetalle{{
			Pago:       models.Pago{ID: 8, PedidoID: 3, UsuarioID: 99, Monto: 35, Referencia: &referencia},
			MetodoPago: models.MetodoPago{ID: 2, Nombre: "Transferencia"},
			Usuario:    &models.PagoUsuario{ID: 99, Nombre: "Empleado"},
		}},
	}
	service := NewPagoService(repo)

	saldo, err := service.ObtenerSaldoCliente(context.Background(), 3, 42)
	if err != nil || saldo.SaldoPendiente != 65 || repo.clienteID != 42 {
		t.Fatalf("saldo=%+v clienteID=%d error=%v", saldo, repo.clienteID, err)
	}
	pagos, err := service.ObtenerHistorialCliente(context.Background(), 3, 42)
	if err != nil || len(pagos) != 1 || pagos[0].ID != 8 || pagos[0].MetodoPago.Nombre != "Transferencia" {
		t.Fatalf("pagos=%+v error=%v", pagos, err)
	}
}

func TestPagoServiceConsultasClienteRechazanIdentidadInvalida(t *testing.T) {
	service := NewPagoService(&fakePagoRepository{})
	if _, err := service.ObtenerSaldoCliente(context.Background(), 3, 0); !errors.Is(err, repository.ErrClienteNoEncontrado) {
		t.Fatalf("ObtenerSaldoCliente() error=%v", err)
	}
	if _, err := service.ObtenerHistorialCliente(context.Background(), 0, 42); !errors.Is(err, repository.ErrPedidoNoEncontrado) {
		t.Fatalf("ObtenerHistorialCliente() error=%v", err)
	}
}
