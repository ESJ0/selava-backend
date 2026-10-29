package service

import (
	"context"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
)

type NotificacionRepository interface {
	ListByCliente(ctx context.Context, clienteID int) (*models.NotificacionesResponse, error)
	Create(ctx context.Context, clienteID int, pedidoID *int, mensaje string) (*models.Notificacion, error)
	MarkReadByCliente(ctx context.Context, notificacionID, clienteID int) (*models.Notificacion, error)
	MarkAllReadByCliente(ctx context.Context, clienteID int) (int64, error)
	DeleteByCliente(ctx context.Context, notificacionID, clienteID int) error
}

type NotificacionService struct {
	repo NotificacionRepository
}

func NewNotificacionService(repo NotificacionRepository) *NotificacionService {
	return &NotificacionService{repo: repo}
}

func (s *NotificacionService) ListarPropias(ctx context.Context, clienteID int) (*models.NotificacionesResponse, error) {
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	return s.repo.ListByCliente(ctx, clienteID)
}

func (s *NotificacionService) Crear(ctx context.Context, clienteID int, pedidoID *int, mensaje string) (*models.Notificacion, error) {
	mensaje = strings.TrimSpace(mensaje)
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	if mensaje == "" {
		return nil, repository.ErrNotificacionNoEncontrada
	}
	return s.repo.Create(ctx, clienteID, pedidoID, mensaje)
}

func (s *NotificacionService) MarcarLeida(ctx context.Context, notificacionID, clienteID int) (*models.Notificacion, error) {
	if notificacionID <= 0 {
		return nil, repository.ErrNotificacionNoEncontrada
	}
	if clienteID <= 0 {
		return nil, repository.ErrClienteNoEncontrado
	}
	return s.repo.MarkReadByCliente(ctx, notificacionID, clienteID)
}

func (s *NotificacionService) MarcarTodasLeidas(ctx context.Context, clienteID int) (int64, error) {
	if clienteID <= 0 {
		return 0, repository.ErrClienteNoEncontrado
	}
	return s.repo.MarkAllReadByCliente(ctx, clienteID)
}

func (s *NotificacionService) Eliminar(ctx context.Context, notificacionID, clienteID int) error {
	if notificacionID <= 0 {
		return repository.ErrNotificacionNoEncontrada
	}
	if clienteID <= 0 {
		return repository.ErrClienteNoEncontrado
	}
	return s.repo.DeleteByCliente(ctx, notificacionID, clienteID)
}
