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
	lastCliente  int

	// err, si no es nil, se devuelve en vez de crear el pedido; simula
	// errores que en produccion vendrian de un constraint de la base de
	// datos (cliente inexistente, tipo de prenda inexistente, etc).
	err error
}

func newFakePedidoRepo() *fakePedidoRepo {
	return &fakePedidoRepo{nextPedidoID: 1, nextPrendaID: 1}
}

func (r *fakePedidoRepo) List(_ context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error) {
	r.lastCliente = filter.ClienteID
	return &models.PedidoListResponse{Pedidos: []models.PedidoResumen{}, Pagina: filter.Pagina, Limite: filter.Limite}, r.err
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

func (r *fakePedidoRepo) UpdateEstado(_ context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error) {
	if r.err != nil {
		return nil, r.err
	}
	r.lastUsuario = usuarioID
	return &models.PedidoEstadoHistorial{
		ID:            1,
		PedidoID:      pedidoID,
		EstadoID:      req.EstadoID,
		UsuarioID:     usuarioID,
		Observaciones: req.Observaciones,
	}, nil
}

func (r *fakePedidoRepo) GetHistorialEstados(_ context.Context, pedidoID int) ([]models.PedidoEstadoHistorialDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	return []models.PedidoEstadoHistorialDetalle{{
		PedidoEstadoHistorial: models.PedidoEstadoHistorial{PedidoID: pedidoID, EstadoID: 1},
		Estado:                models.EstadoPedido{ID: 1, Nombre: "Recibido"},
		Usuario:               models.UsuarioResumen{ID: 7, Nombre: "Ana", Apellido: "Martinez"},
	}}, nil
}

func (r *fakePedidoRepo) GetHistorialEstadosCliente(ctx context.Context, pedidoID, clienteID int) ([]models.PedidoEstadoHistorialDetalle, error) {
	r.lastCliente = clienteID
	return r.GetHistorialEstados(ctx, pedidoID)
}

func (r *fakePedidoRepo) GetDetalle(_ context.Context, pedidoID int) (*models.PedidoDetalle, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &models.PedidoDetalle{
		Pedido:       models.Pedido{ID: pedidoID},
		Cliente:      models.Cliente{ID: 1, Nombre: "Maria"},
		Prendas:      []models.Prenda{{ID: 2, PedidoID: pedidoID}},
		EstadoActual: models.EstadoPedido{ID: 1, Nombre: "Recibido"},
		Pagos:        []models.PagoDetalle{{Pago: models.Pago{ID: 3, PedidoID: pedidoID, Monto: 25}}},
	}, nil
}

func (r *fakePedidoRepo) GetDetalleCliente(ctx context.Context, pedidoID, clienteID int) (*models.PedidoDetalle, error) {
	r.lastCliente = clienteID
	return r.GetDetalle(ctx, pedidoID)
}

func (r *fakePedidoRepo) Cancelar(_ context.Context, pedidoID, usuarioID int) (*models.Pedido, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &models.Pedido{ID: pedidoID, UsuarioID: usuarioID, EstadoActualID: 2, Activo: true}, nil
}

func validPedidoRequest() *models.PedidoCreateRequest {
	return &models.PedidoCreateRequest{
		ClienteID: 1,
		Prendas: []models.PrendaCreateRequest{
			{TipoPrendaID: 1, Cantidad: 2, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: 1}}},
			{TipoPrendaID: 3, Cantidad: 1, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: 2}}},
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

func TestPedidoServiceActualizarEstadoNormalizaYPersiste(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)
	observaciones := "  En proceso  "

	historial, err := service.ActualizarEstado(context.Background(), 4, &models.PedidoEstadoUpdateRequest{
		EstadoID:      2,
		Observaciones: &observaciones,
	}, 7)
	if err != nil {
		t.Fatalf("ActualizarEstado returned error: %v", err)
	}
	if historial.PedidoID != 4 || historial.EstadoID != 2 || historial.UsuarioID != 7 {
		t.Fatalf("unexpected historial: %+v", historial)
	}
	if historial.Observaciones == nil || *historial.Observaciones != "En proceso" {
		t.Fatalf("expected trimmed observations, got %+v", historial.Observaciones)
	}
}

func TestPedidoServiceActualizarEstadoRechazaEstadoInvalido(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	_, err := service.ActualizarEstado(context.Background(), 4, &models.PedidoEstadoUpdateRequest{}, 7)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if !hasFieldError(validationErrors, "estado_id") {
		t.Fatalf("expected error on estado_id, got %+v", validationErrors)
	}
}

func TestPedidoServiceObtenerHistorialEstadosPersisteConsulta(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	historial, err := service.ObtenerHistorialEstados(context.Background(), 4)
	if err != nil {
		t.Fatalf("ObtenerHistorialEstados returned error: %v", err)
	}
	if len(historial) != 1 || historial[0].PedidoID != 4 {
		t.Fatalf("unexpected historial: %+v", historial)
	}
}

func TestPedidoServiceObtenerHistorialEstadosRechazaIDInvalido(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	_, err := service.ObtenerHistorialEstados(context.Background(), 0)
	if !errors.Is(err, repository.ErrPedidoNoEncontrado) {
		t.Fatalf("expected ErrPedidoNoEncontrado, got %v", err)
	}
}

func TestPedidoServiceObtenerDetallePedidoPersisteConsulta(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	detalle, err := service.ObtenerDetallePedido(context.Background(), 4)
	if err != nil {
		t.Fatalf("ObtenerDetallePedido returned error: %v", err)
	}
	if detalle.ID != 4 || detalle.Cliente.ID != 1 || len(detalle.Prendas) != 1 || len(detalle.Pagos) != 1 {
		t.Fatalf("unexpected detalle: %+v", detalle)
	}
}

func TestPedidoServiceObtenerDetallePedidoRechazaIDInvalido(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	_, err := service.ObtenerDetallePedido(context.Background(), 0)
	if !errors.Is(err, repository.ErrPedidoNoEncontrado) {
		t.Fatalf("expected ErrPedidoNoEncontrado, got %v", err)
	}
}

func TestPedidoServiceListaPedidosDelClienteYSanitizaRespuesta(t *testing.T) {
	repo := newFakePedidoRepo()
	service := NewPedidoService(repo)
	result, err := service.ListarPedidosCliente(context.Background(), 42, models.PedidoListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if repo.lastCliente != 42 || result.Pedidos == nil || result.Pagina != 1 || result.Limite != 20 {
		t.Fatalf("resultado=%+v cliente=%d", result, repo.lastCliente)
	}
}

type seguimientoClienteRepo struct {
	*fakePedidoRepo
}

func (r *seguimientoClienteRepo) List(_ context.Context, filter models.PedidoListFilter) (*models.PedidoListResponse, error) {
	r.lastCliente = filter.ClienteID
	return &models.PedidoListResponse{
		Pedidos: []models.PedidoResumen{{
			ID: 7,
			EstadoActual: models.PedidoEstadoResumen{
				ID: 2, Nombre: "En lavado", Orden: 20,
			},
		}},
		Total: 1, Pagina: filter.Pagina, Limite: filter.Limite,
	}, nil
}

func (r *seguimientoClienteRepo) GetDetalleCliente(_ context.Context, pedidoID, clienteID int) (*models.PedidoDetalle, error) {
	r.lastCliente = clienteID
	estadoActual := models.EstadoPedido{ID: 2, Nombre: "En lavado", Codigo: "EN_LAVADO", Orden: 20}
	return &models.PedidoDetalle{
		Pedido:       models.Pedido{ID: pedidoID},
		EstadoActual: estadoActual,
		Prendas:      []models.Prenda{},
		FlujoOperativo: models.FlujoOperativoPedido{
			Estados: []models.EstadoFlujoPedido{
				{EstadoPedido: models.EstadoPedido{ID: 1, Nombre: "Recibido", Codigo: "RECIBIDO", Orden: 10}, Servicios: []models.ServicioEtapa{}, Completado: true},
				{EstadoPedido: estadoActual, Servicios: []models.ServicioEtapa{{ID: 5, Nombre: "Lavado"}}, Completado: true, Actual: true},
				{EstadoPedido: models.EstadoPedido{ID: 3, Nombre: "Finalizado", Codigo: "FINALIZADO", Orden: 60}, Servicios: []models.ServicioEtapa{}},
			},
			SiguienteEstado:    &models.EstadoPedido{ID: 3, Nombre: "Finalizado", Codigo: "FINALIZADO", Orden: 60},
			ProgresoPorcentaje: 66,
		},
	}, nil
}

func (r *seguimientoClienteRepo) GetHistorialEstadosCliente(_ context.Context, pedidoID, clienteID int) ([]models.PedidoEstadoHistorialDetalle, error) {
	r.lastCliente = clienteID
	primero := time.Date(2026, time.October, 27, 8, 30, 0, 0, time.UTC)
	segundo := primero.Add(2 * time.Hour)
	return []models.PedidoEstadoHistorialDetalle{
		{
			PedidoEstadoHistorial: models.PedidoEstadoHistorial{ID: 1, PedidoID: pedidoID, UsuarioID: 91, FechaCambio: primero},
			Estado:                models.EstadoPedido{ID: 1, Nombre: "Recibido", Codigo: "RECIBIDO", Orden: 10},
			Usuario:               models.UsuarioResumen{ID: 91, Nombre: "Dato interno"},
		},
		{
			PedidoEstadoHistorial: models.PedidoEstadoHistorial{ID: 2, PedidoID: pedidoID, UsuarioID: 92, FechaCambio: segundo},
			Estado:                models.EstadoPedido{ID: 2, Nombre: "En lavado", Codigo: "EN_LAVADO", Orden: 20},
			Usuario:               models.UsuarioResumen{ID: 92, Nombre: "Otro dato interno"},
		},
	}, nil
}

func TestPedidoServiceExponeSeguimientoCompletoAlCliente(t *testing.T) {
	repo := &seguimientoClienteRepo{fakePedidoRepo: newFakePedidoRepo()}
	service := NewPedidoService(repo)

	listado, err := service.ListarPedidosCliente(context.Background(), 42, models.PedidoListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listado.Pedidos) != 1 || listado.Pedidos[0].EstadoActual.Nombre != "En lavado" {
		t.Fatalf("listado sin estado oficial: %+v", listado)
	}

	detalle, err := service.ObtenerDetallePedidoCliente(context.Background(), 7, 42)
	if err != nil {
		t.Fatal(err)
	}
	if detalle.EstadoActual.Codigo != "EN_LAVADO" || len(detalle.FlujoOperativo.Estados) != 3 {
		t.Fatalf("detalle sin seguimiento: %+v", detalle)
	}
	etapaLavado := detalle.FlujoOperativo.Estados[1]
	if !etapaLavado.Actual || len(etapaLavado.Servicios) != 1 || etapaLavado.Servicios[0].Nombre != "Lavado" {
		t.Fatalf("etapa aplicable inesperada: %+v", etapaLavado)
	}

	historial, err := service.ObtenerHistorialEstadosCliente(context.Background(), 7, 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(historial) != 2 || historial[0].Estado.Nombre != "Recibido" || historial[1].Estado.Nombre != "En lavado" {
		t.Fatalf("historial inesperado: %+v", historial)
	}
	if !historial[0].FechaCambio.Before(historial[1].FechaCambio) || repo.lastCliente != 42 {
		t.Fatalf("historial fuera de orden o identidad incorrecta: historial=%+v cliente=%d", historial, repo.lastCliente)
	}
}

func TestPedidoServiceRechazaIdentidadClienteInvalida(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())
	if _, err := service.ListarPedidosCliente(context.Background(), 0, models.PedidoListRequest{}); !errors.Is(err, repository.ErrClienteNoEncontrado) {
		t.Fatalf("listado: %v", err)
	}
	if _, err := service.ObtenerDetallePedidoCliente(context.Background(), 1, 0); !errors.Is(err, repository.ErrClienteNoEncontrado) {
		t.Fatalf("detalle: %v", err)
	}
	if _, err := service.ObtenerHistorialEstadosCliente(context.Background(), 1, 0); !errors.Is(err, repository.ErrClienteNoEncontrado) {
		t.Fatalf("historial: %v", err)
	}
}

func TestPedidoServiceCancelarPedidoPersisteCancelacion(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	pedido, err := service.CancelarPedido(context.Background(), 4, 7)
	if err != nil {
		t.Fatalf("CancelarPedido returned error: %v", err)
	}
	if pedido.ID != 4 || pedido.UsuarioID != 7 || pedido.EstadoActualID != 2 {
		t.Fatalf("unexpected pedido cancelado: %+v", pedido)
	}
}

func TestPedidoServiceCancelarPedidoRechazaIDInvalido(t *testing.T) {
	service := NewPedidoService(newFakePedidoRepo())

	_, err := service.CancelarPedido(context.Background(), 0, 7)
	if !errors.Is(err, repository.ErrPedidoNoEncontrado) {
		t.Fatalf("expected ErrPedidoNoEncontrado, got %v", err)
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
			{TipoPrendaID: 0, Cantidad: 0, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: 1}}},
			{TipoPrendaID: 2, Cantidad: -1, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: 1}}},
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
