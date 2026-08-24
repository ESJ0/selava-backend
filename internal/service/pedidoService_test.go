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

type fakePedidoRepo struct {
	nextPedidoID int
	nextPrendaID int
	createCalls  int
	lastUsuario  int

	// err, si no es nil, se devuelve en vez de crear el pedido; simula
	// errores que en produccion vendrian de un constraint de la base de
	// datos (cliente inexistente, tipo de prenda inexistente, etc).
	err error
}

func newFakePedidoRepo() *fakePedidoRepo {
	return &fakePedidoRepo{nextPedidoID: 1, nextPrendaID: 1}
}

func (r *fakePedidoRepo) Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error) {
	r.createCalls++
	r.lastUsuario = usuarioID

	if r.err != nil {
		return nil, r.err
	}

	now := time.Now()
	pedido := models.Pedido{
		ID:                   r.nextPedidoID,
		ClienteID:            req.ClienteID,
		UsuarioID:            usuarioID,
		EstadoActualID:       1,
		FechaRecibido:        now,
		FechaEntregaEstimada: req.FechaEntregaEstimada,
		Observaciones:        req.Observaciones,
		Activo:               true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	r.nextPedidoID++

	prendas := make([]models.Prenda, 0, len(req.Prendas))
	for _, p := range req.Prendas {
		prendas = append(prendas, models.Prenda{
			ID:           r.nextPrendaID,
			PedidoID:     pedido.ID,
			TipoPrendaID: p.TipoPrendaID,
			Descripcion:  p.Descripcion,
			Cantidad:     p.Cantidad,
			Color:        p.Color,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		r.nextPrendaID++
	}

	return &models.PedidoConPrendas{Pedido: pedido, Prendas: prendas}, nil
}

func validPedidoRequest() *models.PedidoCreateRequest {
	return &models.PedidoCreateRequest{
		ClienteID: 1,
		Prendas: []models.PrendaCreateRequest{
			{TipoPrendaID: 1, Cantidad: 2},
			{TipoPrendaID: 3, Cantidad: 1},
		},
	}
}

func TestPedidoServiceCrearPedidoPersistsPedidoYPrendas(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	pedido, err := service.CrearPedido(context.Background(), validPedidoRequest(), 7)
	if err != nil {
		t.Fatalf("CrearPedido returned error: %v", err)
	}

	if pedido.ClienteID != 1 || pedido.UsuarioID != 7 {
		t.Fatalf("unexpected pedido: %+v", pedido)
	}
	if len(pedido.Prendas) != 2 {
		t.Fatalf("expected 2 prendas registradas, got %d", len(pedido.Prendas))
	}
	if pedido.Prendas[0].Cantidad != 2 || pedido.Prendas[1].TipoPrendaID != 3 {
		t.Fatalf("unexpected prendas: %+v", pedido.Prendas)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
	if repo.lastUsuario != 7 {
		t.Fatalf("expected usuarioID 7 to be forwarded, got %d", repo.lastUsuario)
	}
}

func TestPedidoServiceCrearPedidoRejectsClienteIDInvalido(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	req := validPedidoRequest()
	req.ClienteID = 0

	_, err := service.CrearPedido(context.Background(), req, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if !hasFieldError(validationErrors, "cliente_id") {
		t.Fatalf("expected error on cliente_id, got %+v", validationErrors)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestPedidoServiceCrearPedidoRejectsSinPrendas(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	req := validPedidoRequest()
	req.Prendas = nil

	_, err := service.CrearPedido(context.Background(), req, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if !hasFieldError(validationErrors, "prendas") {
		t.Fatalf("expected error on prendas, got %+v", validationErrors)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestPedidoServiceCrearPedidoRejectsCantidadCero(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	req := validPedidoRequest()
	req.Prendas[0].Cantidad = 0

	_, err := service.CrearPedido(context.Background(), req, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if !hasFieldError(validationErrors, "prendas[0].cantidad") {
		t.Fatalf("expected error on prendas[0].cantidad, got %+v", validationErrors)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repo.createCalls)
	}
}

func TestPedidoServiceCrearPedidoRejectsCantidadNegativa(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	req := validPedidoRequest()
	req.Prendas[1].Cantidad = -3

	_, err := service.CrearPedido(context.Background(), req, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if !hasFieldError(validationErrors, "prendas[1].cantidad") {
		t.Fatalf("expected error on prendas[1].cantidad, got %+v", validationErrors)
	}
}

func TestPedidoServiceCrearPedidoRejectsTipoPrendaInvalido(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	req := validPedidoRequest()
	req.Prendas[0].TipoPrendaID = 0

	_, err := service.CrearPedido(context.Background(), req, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if !hasFieldError(validationErrors, "prendas[0].tipo_prenda_id") {
		t.Fatalf("expected error on prendas[0].tipo_prenda_id, got %+v", validationErrors)
	}
}

func TestPedidoServiceCrearPedidoAcumulaErroresDeVariasPrendas(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)

	req := &models.PedidoCreateRequest{
		ClienteID: 0,
		Prendas: []models.PrendaCreateRequest{
			{TipoPrendaID: 0, Cantidad: 0},
			{TipoPrendaID: 2, Cantidad: -1},
		},
	}

	_, err := service.CrearPedido(context.Background(), req, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	// cliente_id + (tipo_prenda_id, cantidad) de la prenda 0 + cantidad de la prenda 1 = 4 errores
	if len(validationErrors) != 4 {
		t.Fatalf("expected 4 validation errors, got %d: %+v", len(validationErrors), validationErrors)
	}
}

func TestPedidoServiceCrearPedidoPropagatesRepositoryError(t *testing.T) {
	repo := newFakePedidoRepo()
	repo.err = repository.ErrClienteNoEncontrado
	service := NewPedidoService(repo)

	_, err := service.CrearPedido(context.Background(), validPedidoRequest(), 7)

	if !errors.Is(err, repository.ErrClienteNoEncontrado) {
		t.Fatalf("expected ErrClienteNoEncontrado, got %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.createCalls)
	}
}

func hasFieldError(errs validator.ValidationErrors, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}
