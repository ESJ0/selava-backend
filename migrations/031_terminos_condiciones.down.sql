ALTER TABLE usuarios
    DROP COLUMN IF EXISTS terminos_version,
    DROP COLUMN IF EXISTS terminos_aceptados_at;
