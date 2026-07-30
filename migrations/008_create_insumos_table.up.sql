CREATE TABLE insumos (
    id             SERIAL PRIMARY KEY,
    nombre         VARCHAR(100)  NOT NULL UNIQUE,
    descripcion    VARCHAR(255),
    unidad_medida  VARCHAR(20)   NOT NULL,
    stock_actual   NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (stock_actual >= 0),
    stock_minimo   NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (stock_minimo >= 0),
    activo         BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP     NOT NULL DEFAULT NOW()
);
