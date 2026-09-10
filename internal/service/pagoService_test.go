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
	created *models.PagoCreateRequest
	err     error
	saldo   *models.SaldoPedido
}

func (r *fakePagoRepository) GetSaldo(_ context.Context, _ int) (*models.SaldoPedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.saldo, nil
}

func (r *fakePagoRepository) Create(_ context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	r.created = req
	return &models.PagoDetalle{Pago: models.Pago{ID: 1, PedidoID: pedidoID, UsuarioID: usuarioID, Monto: req.Monto, Referencia: req.Referencia}}, nil
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
