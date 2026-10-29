# SEL-204: autenticación unificada y rol Recepcionista

## Arquitectura

La migración 028 convierte `usuarios` en la fuente única de credenciales. Agrega
el rol `Cliente`, la relación opcional `usuarios.cliente_id`, metadatos de acceso
y tokens de recuperación. Las filas de `cuentas_cliente` se copian conservando
hash, estado y fechas; la tabla anterior queda solo para permitir rollback.

`POST /api/auth/login` autentica los cuatro roles desde el mismo índice global de
correo y devuelve `{token, usuario, destino}`. Los JWT duran ocho horas e incluyen
rol; los de Cliente también incluyen `cliente_id`. Las contraseñas usan bcrypt y
nunca aparecen en respuestas, logs o almacenamiento del navegador.

## Recuperación

`POST /api/auth/password/forgot` siempre responde 202. El token aleatorio se envía
por correo, pero en base de datos solo se guarda SHA-256. Vence en 30 minutos, una
nueva solicitud invalida tokens anteriores y el cambio lo consume dentro de la
misma transacción. En desarrollo, si `SMTP_HOST` está vacío, el enlace se escribe
en el log local; con SMTP configurado se envía al correo registrado.

## Matriz de permisos

| Módulo | Administrador | Recepcionista | Operario | Cliente |
|---|---|---|---|---|
| Pedidos | Gestión | Gestión | Consulta/flujo operativo existente | Solo propios |
| Clientes | Gestión | Gestión | Bloqueado | Perfil propio |
| Promociones | Gestión | Consulta y aplicación | Bloqueado | Resultado en pedido propio |
| Insumos | Gestión | Gestión | Lectura y movimientos existentes | Bloqueado |
| Tipos de prenda | Gestión | Solo lectura | Bloqueado | Catálogo público |
| Tipos de servicio | Gestión | Solo lectura | Bloqueado | Catálogo público |
| Métodos de pago | Gestión | Bloqueado | Bloqueado | Bloqueado |
| Reportes | Gestión | Bloqueado | Bloqueado | Bloqueado |
| Usuarios | Gestión | Bloqueado | Bloqueado | Bloqueado |

El backend aplica esta matriz aunque alguien escriba una URL manualmente. Los
frontends ocultan los módulos no autorizados y redirigen rutas directas.

## Desarrollo local

`go run ./cmd/seed` crea o actualiza exclusivamente la cuenta administrativa de
prueba `admin@selava.local` con contraseña `Admin123!`. Es un seed local, no una
credencial de producción. El login visible vive en el portal cliente; los roles
internos se transfieren al admin mediante fragmento URL, que el callback consume
y elimina antes de guardar el JWT en `sessionStorage`.
