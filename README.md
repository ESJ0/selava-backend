# selava-backend

## Desarrollo local

Requiere Go 1.25 y Docker. Copia `.env.example` a `.env` y configura la base de datos, `JWT_SECRET`, `SEED_ADMIN_EMAIL` y `SEED_ADMIN_PASSWORD` con valores exclusivamente locales. Docker publica PostgreSQL en `localhost:5433` para no interferir con una instalación local en el puerto estándar.

```powershell
docker compose up -d postgres
.\scripts\migrate.ps1
go run .\cmd\seed
go run .\cmd\api
```

La API queda en `http://localhost:8080` y el health check en `http://localhost:8080/health`. El seed es idempotente y actualiza la contraseña del administrador local desde las variables configuradas. Para detener PostgreSQL sin borrar datos: `docker compose down`.

## Listado de pedidos

`GET /api/pedidos/` requiere JWT de Administrador, Recepcionista u Operario.
Devuelve `{ pedidos, total, pagina, limite }`, incluidos los pedidos cancelados.
Admite `q` (número, nombre o teléfono), `estado_id`, `fecha_desde`, `fecha_hasta`,
`pagina`, `limite` (máximo 100) y `orden` (`recientes` o `antiguos`). Las fechas
de recepción incluyen ambos días en horario de Guatemala; pueden usarse por separado.
La migración 021 añade los índices del listado.

## Catalogo publico (SEL-93)

`GET /api/public/servicios` no requiere JWT ni cabecera `Authorization`. Devuelve
un arreglo de servicios activos, ordenado por ID ascendente (igual que el listado
administrativo), con `id`, `nombre`, `descripcion`, `precio_base` y
`tiempo_estimado_horas`. Los dos campos opcionales pueden ser `null`. Los precios
se consultan en PostgreSQL; no se incluyen timestamps ni datos administrativos.
Un catalogo vacio responde `200 []`. Las rutas `/api/servicios` conservan JWT y
permisos de Administrador/Recepcionista. Contrato completo: `documentos/openapi.yaml`.

Para desarrollar ambos frontends, configura en el `.env` local del backend:
`ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174`. El portal cliente
utiliza el puerto 5174; CORS mantiene una lista explicita de origenes permitidos.

El modelo de identidad para el portal cliente y su relacion opcional 1:1 con
`clientes` se documenta en `documentos/modelo-cuenta-cliente.md`. La migracion
022 crea la tabla `cuentas_cliente`.

El portal dispone de `POST /api/auth/clientes/registro` y
`POST /api/auth/clientes/login`. El registro crea perfil y cuenta en una sola
transaccion; el login devuelve un JWT con `actor=cliente` y `cliente_id`. Estos
tokens no son aceptados por las rutas del panel interno. `GET /api/cliente/me`
es una ruta privada minima para validar la sesion del cliente. El contrato
completo se encuentra en `documentos/openapi.yaml`.

Las garantias y pruebas de aislamiento entre sesiones de clientes se documentan
en `documentos/pruebas-seguridad-clientes.md`. Los tokens del portal no pueden
consultar las rutas administrativas de clientes, pedidos, saldos ni pagos, y la
identidad de `/api/cliente/me` siempre proviene del JWT.

Pruebas: `go test ./...`. Para incluir PostgreSQL, configura
`SELAVA_TEST_DATABASE_URL` apuntando a una base cuyo nombre termine en `_test`.
La infraestructura existente crea un esquema aislado por prueba, aplica las
migraciones reales y lo elimina al terminar; no usa la base de desarrollo.

## Precios por tipo de prenda (SEL-107)

El precio operativo pertenece a `tipos_prenda.precio`. El CRUD exige precio positivo
con hasta dos decimales al crear, y lo devuelve al consultar/listar.
`GET /api/public/tipos-prenda` devuelve solo activos con precio configurado, sin JWT.
`servicios.precio_base` queda deprecado y opcional por compatibilidad.
Las prendas nuevas guardan su tarifa en `prendas.precio_aplicado` y los servicios
no agregan recargos. Los pedidos anteriores conservan todos sus importes.
Consultar [migracion y compatibilidad SEL-107](documentos/sel-107-precios.md).

## Promociones (SEL-109)

La migración aditiva `025_promociones` incorpora el catálogo y los snapshots de
promociones aplicadas. Administrador gestiona `/api/promociones/`; Administrador
y Recepcionista consultan y usan `POST /api/pedidos/cotizacion` para previsualizar
paquetes. El alta de pedido acepta `promociones: [{tipo_prenda_id, promocion_id}]`
y vuelve a validar todos los importes, medidas y vigencias al guardar.
Los detalles administrativos y propios incluyen `promociones_aplicadas`.
Consultar [cálculo exacto, contrato e históricos](documentos/sel-109-promociones.md).
