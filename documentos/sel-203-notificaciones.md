# SEL-203: notificaciones para clientes

## Comportamiento implementado

- Cada cambio confirmado del estado de un pedido crea una notificación para su cliente dentro de la misma transacción.
- El estado `Finalizado` usa un mensaje específico; la cancelación también genera un aviso.
- `pedido_id` es opcional y usa `ON DELETE SET NULL`, por lo que una notificación puede conservarse como aviso general.
- La identidad del propietario siempre procede del JWT de cliente. Los endpoints no aceptan `cliente_id` por URL, query ni body.
- Leer, actualizar y eliminar usan simultáneamente `id` y `cliente_id`; una notificación ajena se responde igual que una inexistente.

## Endpoints del portal

| Método | Ruta | Acción |
| --- | --- | --- |
| `GET` | `/api/cliente/notificaciones/` | Lista el historial y el total no leído |
| `PATCH` | `/api/cliente/notificaciones/{id}/leida` | Marca una como leída |
| `PATCH` | `/api/cliente/notificaciones/leidas` | Marca todas como leídas |
| `DELETE` | `/api/cliente/notificaciones/{id}` | Elimina una notificación propia |

Todos requieren un JWT con `actor=cliente`. La migración requerida es
`027_create_notificaciones.up.sql`.

## Interfaz

El portal vive bajo `/cliente`: usa una sesión distinta del panel interno,
consulta pedidos propios y presenta campana, contador, panel, avisos flotantes,
estados leído/no leído y navegación al pedido relacionado. Actualiza las
notificaciones cada 30 segundos mientras la pestaña está visible y al recuperar
el foco.
