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

Pruebas: `go test ./...`. Para incluir PostgreSQL, configura
`SELAVA_TEST_DATABASE_URL` apuntando a una base cuyo nombre termine en `_test`.
La infraestructura existente crea un esquema aislado por prueba, aplica las
migraciones reales y lo elimina al terminar; no usa la base de desarrollo.
