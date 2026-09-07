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
