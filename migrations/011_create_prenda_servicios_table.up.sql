CREATE TABLE prenda_servicios (
    id              SERIAL PRIMARY KEY,
    prenda_id       INT           NOT NULL REFERENCES prendas(id) ON DELETE CASCADE,
    servicio_id     INT           NOT NULL REFERENCES servicios(id) ON DELETE RESTRICT,
    precio_aplicado NUMERIC(10,2) NOT NULL CHECK (precio_aplicado >= 0),
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_prenda_servicios_prenda_servicio UNIQUE (prenda_id, servicio_id)
);

CREATE INDEX idx_prenda_servicios_prenda_id ON prenda_servicios(prenda_id);
CREATE INDEX idx_prenda_servicios_servicio_id ON prenda_servicios(servicio_id);
