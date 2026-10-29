-- SEL-204: usuarios se convierte en la fuente unica de credenciales.
INSERT INTO roles (id, nombre, descripcion)
VALUES (4, 'Cliente', 'Cliente con acceso al portal publico')
ON CONFLICT (id) DO UPDATE SET nombre = EXCLUDED.nombre, descripcion = EXCLUDED.descripcion, activo = TRUE;

SELECT setval(pg_get_serial_sequence('roles', 'id'), GREATEST((SELECT MAX(id) FROM roles), 1));

ALTER TABLE usuarios
    ADD COLUMN cliente_id INT REFERENCES clientes(id) ON DELETE RESTRICT,
    ADD COLUMN email_verificado_at TIMESTAMP,
    ADD COLUMN ultimo_acceso_at TIMESTAMP;

-- Normaliza los registros historicos antes de validar la nueva restriccion.
UPDATE usuarios SET email = LOWER(BTRIM(email));

ALTER TABLE usuarios
    ADD CONSTRAINT usuarios_cliente_id_key UNIQUE (cliente_id),
    ADD CONSTRAINT chk_usuarios_email_normalizado CHECK (email = LOWER(BTRIM(email)) AND email <> ''),
    ADD CONSTRAINT chk_usuarios_password_hash_no_vacio CHECK (BTRIM(password_hash) <> '');

-- Los correos pasan a ser unicos sin distinguir mayusculas para todos los roles.
CREATE UNIQUE INDEX ux_usuarios_email_normalizado ON usuarios (LOWER(BTRIM(email)));

-- Conserva hash, estado y metadatos de todas las cuentas del portal existentes.
INSERT INTO usuarios (
    rol_id, cliente_id, nombre, apellido, email, password_hash, telefono,
    activo, email_verificado_at, ultimo_acceso_at, created_at, updated_at
)
SELECT
    4, cc.cliente_id, c.nombre, c.apellido, LOWER(BTRIM(cc.email)),
    cc.password_hash, c.telefono, cc.activo AND c.activo,
    cc.email_verificado_at, cc.ultimo_acceso_at, cc.created_at, cc.updated_at
FROM cuentas_cliente cc
JOIN clientes c ON c.id = cc.cliente_id;

CREATE TABLE password_reset_tokens (
    id          BIGSERIAL PRIMARY KEY,
    usuario_id  INT         NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    token_hash  VARCHAR(64) NOT NULL UNIQUE,
    expires_at  TIMESTAMP   NOT NULL,
    used_at     TIMESTAMP,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_password_reset_expiration CHECK (expires_at > created_at)
);

CREATE INDEX idx_password_reset_usuario_vigente
    ON password_reset_tokens (usuario_id, expires_at DESC)
    WHERE used_at IS NULL;

COMMENT ON TABLE cuentas_cliente IS 'Tabla legado conservada para rollback; las credenciales activas viven en usuarios desde SEL-204';
