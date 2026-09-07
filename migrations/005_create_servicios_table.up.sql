CREATE TABLE servicios (
    id                    SERIAL PRIMARY KEY,
    nombre                VARCHAR(100)  NOT NULL UNIQUE,
    descripcion           TEXT,
    precio_base           NUMERIC(10,2) NOT NULL CHECK (precio_base > 0),
    tiempo_estimado_horas INT,
    activo                BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP     NOT NULL DEFAULT NOW()
);
