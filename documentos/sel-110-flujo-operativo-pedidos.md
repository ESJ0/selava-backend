# SEL-110 — Estado y progreso operativo de pedidos

## SEL-110.1 — Auditoría del catálogo anterior

| Estado anterior | Decisión | Motivo |
| --- | --- | --- |
| `Recibido` | Reutilizar | Es el punto de entrada de todo pedido y la única etapa desde la que se permite cancelar. |
| `Rackeado` | Reutilizar | Representa que las prendas terminaron los servicios contratados y quedaron preparadas para retiro. |
| `Entregado` | Renombrar a `Finalizado` | Conserva el mismo ID y las referencias históricas, pero adopta el término terminal solicitado por SEL-110. |
| `Cancelado` | Reutilizar como salida excepcional | No forma parte del avance normal y solo se alcanza mediante el endpoint de cancelación. |

El catálogo anterior no podía representar el trabajo por servicio y permitía saltar cualquier cantidad de estados por el solo hecho de que el `orden` fuera mayor.

## SEL-110.2 — Flujo general

La secuencia queda definida por códigos estables, independientes del texto visible:

1. `RECIBIDO` — obligatoria.
2. `EN_LAVADO` — condicional.
3. `EN_LAVADO_SECO` — condicional.
4. `EN_PLANCHADO` — condicional.
5. `RACKEADO` — obligatoria.
6. `FINALIZADO` — obligatoria y terminal.

`CANCELADO` es terminal, pero no pertenece a la secuencia normal. Las decenas usadas en `orden` permiten agregar etapas futuras entre las existentes sin renumerar todo el catálogo.

## SEL-110.3 — Etapas por servicios contratados

La relación muchos-a-muchos `estado_pedido_servicios` determina qué etapas condicionales aplican. El flujo de cada pedido se calcula con los servicios realmente asociados a sus prendas, no con el catálogo vigente ni con una selección enviada por el navegador.

| Servicio auditado | Etapa habilitada |
| --- | --- |
| `Lavado` | `EN_LAVADO` |
| `Lavado en seco` | `EN_LAVADO_SECO` |
| `Planchado` | `EN_PLANCHADO` |

Las etapas `RECIBIDO`, `RACKEADO` y `FINALIZADO` siempre aplican. Si varias prendas contratan el mismo servicio, la etapa aparece una sola vez y expone los servicios que la originaron. Un trigger mantiene las relaciones de los tres servicios auditados tanto para bases ya pobladas como para instalaciones donde el seed se ejecuta después de migrar.

Los servicios futuros no se infieren por coincidencias parciales: deben relacionarse explícitamente con una etapa para evitar que un nombre ambiguo altere el flujo de producción.

## SEL-110.4 — Transiciones válidas

- Solo se acepta la etapa inmediatamente siguiente del flujo calculado para el pedido.
- No se permiten saltos, repeticiones ni retrocesos.
- Una etapa que corresponde a un servicio no contratado no puede seleccionarse.
- `Cancelado` solo se establece con `PUT /api/pedidos/{id}/cancelar` y únicamente desde `Recibido`.
- `Finalizado` y `Cancelado` son terminales; no admiten nuevas transiciones ni modificaciones de prendas o servicios.
- Cada cambio actualiza el pedido y crea el historial en la misma transacción y bajo bloqueo del pedido.
- Al alcanzar `Finalizado`, el backend registra `fecha_entrega_real`.
- La creación registra `Recibido` como primera entrada del historial.
- Solo `Administrador` y `Recepcionista` pueden ejecutar el cambio; `Operario` conserva acceso de lectura al catálogo, detalle e historial, pero recibe `403` al intentar actualizar.

## Contrato para Admin

`GET /api/pedidos/{id}` incluye `flujo_operativo` con:

- `estados`: solo las etapas aplicables, sus servicios causantes y sus marcas `completado`/`actual`;
- `siguiente_estado`: el único destino válido, si existe;
- `progreso_porcentaje`: etapas alcanzadas sobre el total aplicable;
- `cancelado`: distingue un cierre cancelado de uno finalizado.

El Admin usa este contrato como fuente de verdad: muestra el progreso completo y solo permite avanzar al `siguiente_estado`. El backend vuelve a validar la transición para impedir saltos mediante llamadas directas o interfaces desactualizadas.

