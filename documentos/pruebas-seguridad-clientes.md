# Validación de seguridad del portal de clientes

## Límite de confianza

La identidad del cliente se obtiene exclusivamente del claim `cliente_id` de
un JWT con `actor=cliente`. El backend ignora identificadores enviados mediante
query parameters o cabeceras personalizadas para decidir la identidad de la
sesión.

Los tokens del portal y del panel interno no son intercambiables:

- `AuthenticateCliente` solo acepta `actor=cliente`.
- `Authenticate` solo acepta `actor=usuario` junto con `usuario_id` y `rol_id`.
- Las rutas administrativas de clientes, pedidos, saldos y pagos usan
  `Authenticate`; por eso un token de cliente recibe `401` antes de ejecutar el
  controlador o consultar PostgreSQL.

Actualmente el único recurso privado del portal es `GET /api/cliente/me`. Los
endpoints administrativos de pedidos y pagos no se reutilizan para el portal.
Cuando se incorporen rutas como “Mis pedidos”, cada consulta deberá recibir el
`cliente_id` del contexto autenticado y filtrarlo en PostgreSQL. No se deberá
aceptar un `cliente_id` procedente de la URL, query o body. Un recurso ajeno debe
responder como no encontrado para no revelar su existencia.

## Matriz automatizada

| Escenario | Resultado esperado |
| --- | --- |
| Registro válido | `201`, perfil y cuenta relacionados, sin hash en JSON |
| Registro con el mismo correo en mayúsculas | `409` |
| Registro con campos inválidos | `422` |
| Login con contraseña incorrecta | `401` sin emitir token |
| Login válido | JWT con `actor=cliente` y `cliente_id` correcto |
| Ruta privada sin token | `401` |
| `/api/cliente/me` con token válido | Identidad tomada del JWT |
| Intento de cambiar identidad por query o cabecera | Se ignora y prevalece el JWT |
| Token de cliente en `/api/clientes/{id}` | `401` |
| Token de cliente en detalle, pagos o saldo administrativo | `401` |
| Token de usuario interno en `/api/cliente/me` | `401` |

Las pruebas HTTP integrales se encuentran en
`internal/routes/cliente_auth_flow_test.go` y
`internal/routes/cliente_data_isolation_test.go`. Usan las capas reales de
validación, bcrypt, JWT, controladores, middleware y rutas; solo PostgreSQL se
reemplaza por un repositorio en memoria aislado para la prueba.
