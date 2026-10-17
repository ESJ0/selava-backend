CREATE TABLE cuentas_cliente (
    id                   SERIAL PRIMARY KEY,
    cliente_id           INT          NOT NULL UNIQUE REFERENCES clientes(id) ON DELETE RESTRICT,
    email                VARCHAR(150) NOT NULL,
    password_hash        VARCHAR(255) NOT NULL,
    activo               BOOLEAN      NOT NULL DEFAULT TRUE,
    email_verificado_at  TIMESTAMP,
    ultimo_acceso_at     TIMESTAMP,
    created_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_cuentas_cliente_email_normalizado
        CHECK (email = LOWER(BTRIM(email)) AND email <> ''),
    CONSTRAINT chk_cuentas_cliente_password_hash_no_vacio
        CHECK (BTRIM(password_hash) <> '')
);

-- El correo de acceso es insensible a mayusculas. La restriccion CHECK obliga
-- a persistirlo normalizado y este indice garantiza su unicidad global.
CREATE UNIQUE INDEX ux_cuentas_cliente_email
    ON cuentas_cliente (LOWER(BTRIM(email)));
