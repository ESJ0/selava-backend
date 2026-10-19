package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type PagoRepository interface {
	Create(ctx context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error)
	GetSaldo(ctx context.Context, pedidoID int) (*models.SaldoPedido, error)
	GetSaldoCliente(ctx context.Context, pedidoID, clienteID int) (*models.SaldoPedido, error)
	ListByPedido(ctx context.Context, pedidoID int) ([]models.PagoDetalle, error)
	ListByPedidoCliente(ctx context.Context, pedidoID, clienteID int) ([]models.PagoDetalle, error)
}

type PagoService struct{ repo PagoRepository }

func NewPagoService(repo PagoRepository) *PagoService { return &PagoService{repo: repo} }

func (s *PagoService) ObtenerSaldo(ctx context.Context, pedidoID int) (*models.SaldoPedido, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	return s.repo.GetSaldo(ctx, pedidoID)
}

func (s *PagoService) ObtenerHistorial(ctx context.Context, pedidoID int) ([]models.PagoDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	return s.repo.ListByPedido(ctx, pedidoID)
}

func (s *PagoService) ObtenerSaldoCliente(ctx context.Context, pedidoID, clienteID int) (*models.SaldoPedido, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	return s.repo.GetSaldoCliente(ctx, pedidoID, clienteID)
}

func (s *PagoService) ObtenerHistorialCliente(ctx context.Context, pedidoID, clienteID int) ([]models.PagoPropioDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	pagos, err := s.repo.ListByPedidoCliente(ctx, pedidoID, clienteID)
	if err != nil {
		return nil, err
	}
	result := make([]models.PagoPropioDetalle, 0, len(pagos))
	for _, pago := range pagos {
		result = append(result, models.PagoPropioDetalle{
			ID: pago.ID, PedidoID: pago.PedidoID, Monto: pago.Monto,
			Referencia: pago.Referencia, FechaPago: pago.FechaPago, MetodoPago: pago.MetodoPago,
		})
	}
	return result, nil
}

func (s *PagoService) RegistrarPago(ctx context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if usuarioID <= 0 {
		return nil, repository.ErrUsuarioNoEncontrado
	}
	if req.Referencia != nil {
		referencia := strings.TrimSpace(*req.Referencia)
		if referencia == "" {
			req.Referencia = nil
		} else {
			req.Referencia = &referencia
		}
	}
	if errs := validator.ValidatePagoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, pedidoID, req, usuarioID)
}
