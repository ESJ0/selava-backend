DROP TABLE IF EXISTS password_reset_tokens;

-- Copia al modelo legado cualquier cliente creado despues de la migracion.
INSERT INTO cuentas_cliente (
    cliente_id, email, password_hash, activo, email_verificado_at,
    ultimo_acceso_at, created_at, updated_at
)
SELECT
    u.cliente_id, u.email, u.password_hash, u.activo, u.email_verificado_at,
    u.ultimo_acceso_at, u.created_at, u.updated_at
FROM usuarios u
WHERE u.rol_id = 4 AND u.cliente_id IS NOT NULL
ON CONFLICT (cliente_id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    activo = EXCLUDED.activo,
    email_verificado_at = EXCLUDED.email_verificado_at,
    ultimo_acceso_at = EXCLUDED.ultimo_acceso_at,
    updated_at = EXCLUDED.updated_at;

DELETE FROM usuarios WHERE rol_id = 4;
DROP INDEX IF EXISTS ux_usuarios_email_normalizado;
ALTER TABLE usuarios
    DROP CONSTRAINT IF EXISTS chk_usuarios_password_hash_no_vacio,
    DROP CONSTRAINT IF EXISTS chk_usuarios_email_normalizado,
    DROP CONSTRAINT IF EXISTS usuarios_cliente_id_key,
    DROP COLUMN IF EXISTS ultimo_acceso_at,
    DROP COLUMN IF EXISTS email_verificado_at,
    DROP COLUMN IF EXISTS cliente_id;
DELETE FROM roles WHERE id = 4 AND nombre = 'Cliente';
COMMENT ON TABLE cuentas_cliente IS NULL;
