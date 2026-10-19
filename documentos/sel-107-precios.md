# SEL-107: precio por tipo de prenda

## Modelo y transición

`tipos_prenda.precio` es la tarifa operativa en quetzales. Es `NUMERIC` con
restricciones `> 0`, `<= 99999999.99` y `precio = ROUND(precio, 2)`. Se usa
NUMERIC sin escala declarada para **rechazar**, en vez de redondear, entradas
con fracciones de centavo. El API valida las mismas reglas y devuelve 422 con
`{"errores":[{"field":"precio","message":"..."}]}`.

La migración 023 no puede inferir la tarifa de cada tipo desde un catálogo
independiente de servicios: no existe una correspondencia única. Los tipos
anteriores quedan con `precio = NULL`, que significa pendiente de configurar.
Un trigger exige precio a todos los tipos nuevos e impide borrar una tarifa ya
configurada. Las actualizaciones parciales de un tipo anterior pueden conservar
NULL durante la transición. El admin muestra el pendiente y permite asignar su
precio; no permite seleccionarlo para nuevas ventas. El API rechaza un nuevo
cobro de ese tipo con 422 (`tipo_prenda_id`), sin guardar pedidos incompletos.

El CRUD recibe `precio` al crear y opcionalmente al editar. Listar y obtener
devuelven el precio, o NULL únicamente para los registros anteriores pendientes.
Nombre, descripción, activo y los demás campos conservan su comportamiento.

## Cálculo e históricos

Antes, cada asociación copiaba `servicios.precio_base` a
`prenda_servicios.precio_aplicado` y el total sumaba esos valores por cantidad.

Ahora ambos caminos de alta (pedido con prendas y agregar prendas a un pedido)
comparten `insertarPrendaConPrecio`: leen la tarifa activa con `FOR SHARE` y la
guardan en **`prendas.precio_aplicado`** dentro de la misma transacción. El total
usa esa instantánea una vez por prenda, multiplicada por la cantidad existente.
Agregar varios servicios no multiplica el precio. Por ejemplo, Camisa Q25 con
Lavado y Planchado cuesta Q25; dos camisas cuestan Q50. Las asociaciones nuevas
guardan cero como importe de servicio, manteniendo sus identificadores y relaciones.
El precio nunca se toma del request de prendas ni del servicio.

La migración no actualiza `pedidos.total`, `prenda_servicios.precio_aplicado`,
pagos ni cantidades existentes. `prendas.precio_aplicado = NULL` identifica las
prendas anteriores: incluso si se recalcula un pedido por una operación explícita,
esas filas siguen usando **sus importes históricos congelados por servicio**,
nunca tarifas actuales. Así se pueden sumar prendas antiguas y nuevas en el mismo
pedido. Consultar o modificar precios del catálogo no recalcula pedidos.

El detalle expone `prendas.precio_aplicado` para que el admin muestre el importe
congelado y las asociaciones sin recargos. En registros anteriores conserva la
presentación de sus precios aplicados por servicio. La tarifa actual del catálogo
puede aparecer en el tipo relacionado, pero no se usa para presentar el cobro pasado.

## Servicios y compatibilidad

Se conserva `servicios.precio_base` y sus valores antiguos. La columna admite
NULL para nuevas altas sin tarifa. Los contratos anteriores siguen aceptando y
devolviendo `precio_base`, marcado deprecado en OpenAPI. Un alta que lo omite
(o envía cero) guarda NULL; las respuestas conservan el tipo numérico y devuelven
cero cuando no hay tarifa heredada. Un precio heredado negativo sigue siendo inválido.
Ningún cálculo de prendas nuevas lee esta columna. El admin elimina el precio
de formulario, listado y selector de servicios; conserva los otros campos.

`GET /api/public/servicios` permanece compatible para consumidores anteriores,
pero su precio está deprecado. La landing usa exclusivamente
`GET /api/public/tipos-prenda`, sin JWT ni rutas administrativas, que devuelve
`id`, `nombre`, `descripcion` y `precio` de tipos activos con tarifa configurada,
ordenados por ID. El resultado vacío es `[]`.

## Aplicación y reversión

1. Respaldar la base y aplicar `023_garment_type_pricing.up.sql` en una transacción
   antes de iniciar el backend SEL-107. Coordinar la actualización del backend y
   los dos frontends para que las altas de tipos incluyan precio.
2. Configurar las tarifas reales de los tipos anteriores desde el admin. No hay
   precios inventados ni copia automática de tarifas de servicios.
3. Comprobar el catálogo público y una nueva venta con varios servicios.

El down es válido antes de registrar prendas con instantánea o servicios sin tarifa.
Después rechaza la reversión para evitar pérdida de importes. En ese caso se debe
usar una migración correctiva que conserve los datos, nunca borrar instantáneas.

El seed de desarrollo también usa el modelo nuevo: servicios sin tarifa y tipos
de ejemplo a Q25 (solo datos locales de demostración). No sobrescribe precios
configurados ni asigna precios a tipos anteriores que estén pendientes.

## Validación

Las pruebas de repositorio usan `SELAVA_TEST_DATABASE_URL`, que debe terminar en
`_test`, y esquemas aislados con migraciones reales. Cubren CRUD, restricciones
SQL, catálogo público, múltiples servicios, ambas vías de creación de prendas,
cambios de tarifa, rollback, migración desde 022 con un pedido existente y bloqueo
de una reversión destructiva. No usan `.env` ni datos de desarrollo.

La base `develop` traía un doble de pruebas sin `GetClienteActivoByID` y aserciones
de `/api/cliente/me` que buscaban `cliente_id` aunque el contrato real devuelve `id`.
Se completa exclusivamente ese doble y se corrigen esas aserciones para poder
ejecutar la suite completa, sin modificar autenticación ni aislamiento en producción.

SEL-108 (modalidades UNIDAD/LIBRA/METRO) y SEL-109 (promociones) no se implementan.
La tarifa pertenece al tipo; la medición y futuras modalidades quedan separadas.
