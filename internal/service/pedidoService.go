package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

// PedidoRepository es la interfaz que necesita el service para poder
// probarse con un fake, en vez de depender de *repository.PedidoRepository.
type PedidoRepository interface {
	List(ctx context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error)
	Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error)
	UpdateEstado(ctx context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error)
	GetHistorialEstados(ctx context.Context, pedidoID int) ([]models.PedidoEstadoHistorialDetalle, error)
	GetHistorialEstadosCliente(ctx context.Context, pedidoID, clienteID int) ([]models.PedidoEstadoHistorialDetalle, error)
	GetDetalle(ctx context.Context, pedidoID int) (*models.PedidoDetalle, error)
	GetDetalleCliente(ctx context.Context, pedidoID, clienteID int) (*models.PedidoDetalle, error)
	Cancelar(ctx context.Context, pedidoID int, usuarioID int) (*models.Pedido, error)
}

type PedidoService struct {
	repo PedidoRepository
}

func NewPedidoService(repo PedidoRepository) *PedidoService {
	return &PedidoService{repo: repo}
}

func (s *PedidoService) ListarPedidos(ctx context.Context, req models.PedidoListRequest) (*models.PedidoListResponse, error) {
	filter, errs := validator.ParsePedidoList(req)
	if errs.HasErrors() {
		return nil, errs
	}
	return s.repo.List(ctx, filter)
}

func (s *PedidoService) ListarPedidosCliente(ctx context.Context, clienteID int, req models.PedidoListRequest) (*models.PedidosPropiosResponse, error) {
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	filter, errs := validator.ParsePedidoList(req)
	if errs.HasErrors() {
		return nil, errs
	}
	filter.ClienteID = clienteID
	result, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	response := &models.PedidosPropiosResponse{
		Pedidos: make([]models.PedidoPropioResumen, 0, len(result.Pedidos)),
		Total:   result.Total, Pagina: result.Pagina, Limite: result.Limite,
	}
	for _, pedido := range result.Pedidos {
		response.Pedidos = append(response.Pedidos, models.PedidoPropioResumen{
			ID: pedido.ID, FechaRecibido: pedido.FechaRecibido,
			FechaEntregaEstimada: pedido.FechaEntregaEstimada, Total: pedido.Total,
			EstadoActual: pedido.EstadoActual,
		})
	}
	return response, nil
}

func (s *PedidoService) CrearPedido(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error) {
	for i := range req.Prendas {
		if req.Prendas[i].Color != nil {
			value := strings.TrimSpace(*req.Prendas[i].Color)
			req.Prendas[i].Color = &value
		}
		if req.Prendas[i].Descripcion != nil {
			value := strings.TrimSpace(*req.Prendas[i].Descripcion)
			req.Prendas[i].Descripcion = &value
		}
	}
	if errs := validator.ValidatePedidoCreate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.Create(ctx, req, usuarioID)
}

func (s *PedidoService) ActualizarEstado(ctx context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if usuarioID <= 0 {
		return nil, repository.ErrUsuarioNoEncontrado
	}
	if req.Observaciones != nil {
		value := strings.TrimSpace(*req.Observaciones)
		req.Observaciones = &value
	}
	if errs := validator.ValidatePedidoEstadoUpdate(req); errs.HasErrors() {
		return nil, errs
	}
	return s.repo.UpdateEstado(ctx, pedidoID, req, usuarioID)
}

func (s *PedidoService) ObtenerHistorialEstados(ctx context.Context, pedidoID int) ([]models.PedidoEstadoHistorialDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	return s.repo.GetHistorialEstados(ctx, pedidoID)
}

func (s *PedidoService) ObtenerHistorialEstadosCliente(ctx context.Context, pedidoID, clienteID int) ([]models.PedidoPropioHistorial, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	historial, err := s.repo.GetHistorialEstadosCliente(ctx, pedidoID, clienteID)
	if err != nil {
		return nil, err
	}
	result := make([]models.PedidoPropioHistorial, 0, len(historial))
	for _, movimiento := range historial {
		result = append(result, models.PedidoPropioHistorial{
			ID: movimiento.ID, PedidoID: movimiento.PedidoID, Estado: movimiento.Estado,
			FechaCambio: movimiento.FechaCambio, Observaciones: movimiento.Observaciones,
		})
	}
	return result, nil
}

func (s *PedidoService) ObtenerDetallePedido(ctx context.Context, pedidoID int) (*models.PedidoDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	return s.repo.GetDetalle(ctx, pedidoID)
}

func (s *PedidoService) ObtenerDetallePedidoCliente(ctx context.Context, pedidoID, clienteID int) (*models.PedidoPropioDetalle, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	detalle, err := s.repo.GetDetalleCliente(ctx, pedidoID, clienteID)
	if err != nil {
		return nil, err
	}
	return &models.PedidoPropioDetalle{
		ID: detalle.ID, FechaRecibido: detalle.FechaRecibido,
		FechaEntregaEstimada: detalle.FechaEntregaEstimada, FechaEntregaReal: detalle.FechaEntregaReal,
		Total: detalle.Total, Observaciones: detalle.Observaciones, Prendas: detalle.Prendas,
		EstadoActual:         detalle.EstadoActual,
		PromocionesAplicadas: detalle.PromocionesAplicadas,
	}, nil
}

func (s *PedidoService) CancelarPedido(ctx context.Context, pedidoID, usuarioID int) (*models.Pedido, error) {
	if pedidoID <= 0 {
		return nil, repository.ErrPedidoNoEncontrado
	}
	if usuarioID <= 0 {
		return nil, repository.ErrUsuarioNoEncontrado
	}
	return s.repo.Cancelar(ctx, pedidoID, usuarioID)
}
