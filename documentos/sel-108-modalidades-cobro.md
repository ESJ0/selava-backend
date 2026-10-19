# SEL-108: cobro por unidad, libra o metro

## Modelo

`tipos_prenda.modalidad_cobro` define `UNIDAD`, `LIBRA` o `METRO`. La migración
024 asigna `UNIDAD` a los tipos existentes y configura Canasta como `LIBRA` y
Alfombra como `METRO` cuando esos datos existen. El seed local crea ambos tipos
con una tarifa demostrativa; la tarifa real continúa siendo responsabilidad del
administrador.

`prendas.cantidad` conserva su nombre para mantener compatible el contrato, pero
ahora representa la medida cobrada y admite hasta tres decimales. Cada alta copia
la modalidad a `prendas.modalidad_cobro_aplicada` y el precio vigente a
`prendas.precio_aplicado`. Esas instantáneas protegen el histórico frente a
cambios posteriores del catálogo.

## Reglas y cálculo

- `UNIDAD`: cantidad positiva, entera y, en el formulario, máximo 50 por grupo.
- `LIBRA` y `METRO`: medida positiva con hasta tres decimales.
- Cada subtotal es `ROUND(cantidad * precio_aplicado, 2)` en PostgreSQL; el total
  suma esos centavos. El admin usa el mismo redondeo decimal por línea, sin
  errores binarios de `toFixed`. Los importes que exceden el límite de la base
  se rechazan con 422 y rollback de la operación completa.
- El request no contiene precio, subtotal ni total. Los dos caminos de alta de
  prendas consultan precio y modalidad dentro de su transacción y el backend
  recalcula `pedidos.total` antes de confirmar.
- Canasta y Alfombra son configuración de datos; el código de cálculo nunca usa
  el nombre del tipo.

Las prendas anteriores a SEL-107 conservan su cálculo histórico por servicios.
La migración 024 solo agrega su modalidad inicial y no cambia importes ni pagos.
Toda prenda anterior recibe `UNIDAD`: no se reinterpreta una Canasta o Alfombra
histórica según la nueva modalidad del catálogo. El down rechaza la pérdida de
cualquier modalidad LIBRA/METRO, incluso si su medida es entera. El seed no
sobrescribe modalidades configuradas por el administrador.

## Frontends

El panel administrativo mantiene el wizard existente. El campo y su validación
cambian dinámicamente entre Cantidad, Peso en libras y Metros. El resumen y el
detalle muestran medida, unidad, tarifa congelada y subtotal. El portal cliente
requiere indicar «por unidad/libra/metro» en el catálogo público y presentar la
modalidad congelada en el detalle. En la revisión no se encontró una rama cliente
SEL-108; ese ajuste queda pendiente por instrucción expresa del usuario. Los PRs
de backend/admin no se deben integrar mientras esa incompatibilidad permanezca.
