# Endpoints de pedidos — Sprint 3

Base local: `http://localhost:8080`. Todas las rutas requieren `Authorization: Bearer <JWT>`. Los IDs de ruta son enteros positivos. Las respuestas de error usan `{"error":"mensaje"}`; las validaciones usan `{"errores":[{"field":"campo","message":"detalle"}]}`.

## Servicios por prenda

### Agregar servicio

`POST /api/prendas/{prendaID}/servicios`

Roles: Administrador, Recepcionista. El backend toma `precio_aplicado` de `servicios.precio_base`; ignora cualquier precio enviado y recalcula `pedidos.total` dentro de la misma transacción.

```json
// Request
{"servicio_id": 3}

// 201
{"id": 15, "prenda_id": 10, "servicio_id": 3, "precio_aplicado": 25.00, "created_at": "2026-09-07T16:30:00Z", "updated_at": "2026-09-07T16:30:00Z"}
```

Errores: 400 ID/JSON inválido; 401 sin JWT; 403 rol no autorizado; 404 prenda o servicio activo inexistente; 409 asociación duplicada o pedido Entregado; 422 `servicio_id` inválido; 500 error interno.

### Quitar servicio

`DELETE /api/prendas/{prendaID}/servicios/{servicioID}`

Roles: Administrador, Recepcionista. Elimina la relación y recalcula el total en una transacción.

Respuesta: `204 No Content` sin cuerpo.

Errores: 400 ID inválido; 401; 403; 404 prenda/asociación inexistente; 409 pedido Entregado; 500.

## Estados

### Catálogo de estados

`GET /api/estados-pedido/`

Roles: Administrador, Recepcionista, Operario.

```json
[
  {"id":1,"nombre":"Recibido","orden":1},
  {"id":3,"nombre":"Rackeado","orden":2},
  {"id":4,"nombre":"Entregado","orden":3},
  {"id":2,"nombre":"Cancelado","orden":99}
]
```

Errores: 401; 500.

### Actualizar estado

`PUT /api/pedidos/{pedidoID}/estado`

Roles: Administrador, Recepcionista, Operario. Actualiza el pedido y registra usuario, fecha y observaciones en el historial dentro de una transacción. Permite avanzar a cualquier orden superior; no exige pasar por todos los estados. `Cancelado` solo se establece mediante el endpoint de cancelación.

```json
// Request
{"estado_id": 3, "observaciones": "Pedido listo para entrega"}

// 200
{"id":12,"pedido_id":42,"estado_id":3,"usuario_id":7,"fecha_cambio":"2026-09-07T16:30:00Z","observaciones":"Pedido listo para entrega","created_at":"2026-09-07T16:30:00Z"}
```

Errores: 400 ID/JSON inválido; 401; 403; 404 pedido/estado inexistente; 409 retroceso, pedido Entregado/Cancelado o intento de usar Cancelado; 422 campos inválidos (`observaciones` máximo 255 bytes); 500.

### Consultar historial

`GET /api/pedidos/{pedidoID}/historial-estados`

Roles: Administrador, Recepcionista, Operario. Devuelve entradas desde la más antigua a la más reciente, con estado y usuario resueltos.

```json
[
  {
    "id":12,"pedido_id":42,"estado_id":3,"usuario_id":7,
    "fecha_cambio":"2026-09-07T16:30:00Z","observaciones":"Pedido listo",
    "estado":{"id":3,"nombre":"Rackeado","orden":2},
    "usuario":{"id":7,"nombre":"Ana","apellido":"Martinez"}
  }
]
```

Errores: 400 ID inválido; 401; 404 pedido inexistente/inactivo; 500.

## Detalle completo

`GET /api/pedidos/{pedidoID}`

Roles: Administrador, Recepcionista, Operario. Incluye cliente, prendas, tipo de prenda, servicios y precios aplicados, total, estado actual y pagos.

```json
{
  "id":42,"cliente_id":4,"usuario_id":7,"estado_actual_id":1,"total":50,
  "cliente":{"id":4,"nombre":"Cliente","apellido":"Demo","telefono":"55550000","activo":true},
  "prendas":[{"id":10,"pedido_id":42,"tipo_prenda_id":2,"cantidad":2,"tipo_prenda":{"id":2,"nombre":"Camisa","activo":true},"servicios":[{"id":15,"prenda_id":10,"servicio_id":3,"precio_aplicado":25,"servicio":{"id":3,"nombre":"Lavado","precio_base":25,"activo":true}}]}],
  "estado_actual":{"id":1,"nombre":"Recibido","orden":1},
  "pagos":[]
}
```

Errores: 400 ID inválido; 401; 404 pedido inexistente/inactivo; 500.

## Cancelación

`PUT /api/pedidos/{pedidoID}/cancelar`

Roles: Administrador, Recepcionista. No recibe cuerpo. Solo un pedido `Recibido` puede cancelarse; la actualización y su entrada de historial son atómicas.

```json
// 200
{"id":42,"cliente_id":4,"usuario_id":7,"estado_actual_id":2,"total":50,"activo":true,"updated_at":"2026-09-07T16:40:00Z"}
```

Errores: 400 ID inválido; 401; 403 (incluye Operario); 404 pedido inexistente/inactivo; 409 cualquier estado distinto de Recibido; 500 si no está configurado el estado Cancelado o falla el servidor.
