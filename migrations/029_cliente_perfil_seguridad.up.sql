ALTER TABLE clientes
    ADD COLUMN foto_perfil BYTEA,
    ADD COLUMN foto_perfil_tipo VARCHAR(50);

ALTER TABLE clientes
    ADD CONSTRAINT chk_clientes_foto_perfil_completa CHECK (
        (foto_perfil IS NULL AND foto_perfil_tipo IS NULL)
        OR (foto_perfil IS NOT NULL AND foto_perfil_tipo IN ('image/jpeg', 'image/png', 'image/webp'))
    );
