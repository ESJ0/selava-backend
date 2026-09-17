package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/validator"
)

type fakeMovimientoInventarioRepo struct {
	receivedReq  *models.MovimientoInventarioCreateRequest
	receivedUser int
	createCalls  int
}

func (r *fakeMovimientoInventarioRepo) Create(ctx context.Context, req *models.MovimientoInventarioCreateRequest, usuarioID int) (*models.MovimientoInventario, error) {
	r.receivedReq = req
	r.receivedUser = usuarioID
	r.createCalls++
	return &models.MovimientoInventario{
		ID: 1, InsumoID: req.InsumoID, UsuarioID: usuarioID,
		TipoMovimiento: req.TipoMovimiento, Cantidad: req.Cantidad,
		Motivo: req.Motivo, FechaMovimiento: time.Now(), CreatedAt: time.Now(),
	}, nil
}

func TestMovimientoInventarioServiceRegistraYNormaliza(t *testing.T) {
	repo := &fakeMovimientoInventarioRepo{}
	service := NewMovimientoInventarioService(repo)
	motivo := "  Compra de insumo  "

	movimiento, err := service.RegistrarMovimiento(context.Background(), &models.MovimientoInventarioCreateRequest{
		InsumoID: 4, TipoMovimiento: "entrada", Cantidad: 12.5, Motivo: &motivo,
	}, 7)
	if err != nil {
		t.Fatalf("RegistrarMovimiento returned error: %v", err)
	}
	if movimiento.UsuarioID != 7 || repo.receivedUser != 7 {
		t.Fatalf("expected authenticated user 7, got movement=%d repo=%d", movimiento.UsuarioID, repo.receivedUser)
	}
	if repo.receivedReq.Motivo == nil || *repo.receivedReq.Motivo != "Compra de insumo" {
		t.Fatalf("expected trimmed motivo, got %+v", repo.receivedReq.Motivo)
	}
}

func TestMovimientoInventarioServiceRechazaDatosInvalidos(t *testing.T) {
	repo := &fakeMovimientoInventarioRepo{}
	service := NewMovimientoInventarioService(repo)

	_, err := service.RegistrarMovimiento(context.Background(), &models.MovimientoInventarioCreateRequest{
		InsumoID: 4, TipoMovimiento: "ajuste", Cantidad: 0,
	}, 7)
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestMovimientoInventarioServiceRechazaUsuarioInvalido(t *testing.T) {
	repo := &fakeMovimientoInventarioRepo{}
	service := NewMovimientoInventarioService(repo)

	_, err := service.RegistrarMovimiento(context.Background(), &models.MovimientoInventarioCreateRequest{
		InsumoID: 4, TipoMovimiento: "salida", Cantidad: 1,
	}, 0)
	if !errors.Is(err, repository.ErrUsuarioNoEncontrado) {
		t.Fatalf("expected ErrUsuarioNoEncontrado, got %v", err)
	}
}
