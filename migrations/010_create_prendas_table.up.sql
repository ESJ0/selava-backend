CREATE TABLE prendas (
    id             SERIAL PRIMARY KEY,
    pedido_id      INT         NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    tipo_prenda_id INT         NOT NULL REFERENCES tipos_prenda(id) ON DELETE RESTRICT,
    descripcion    VARCHAR(255),
    cantidad       INT         NOT NULL DEFAULT 1 CHECK (cantidad > 0),
    color          VARCHAR(30),
    created_at     TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prendas_pedido_id ON prendas(pedido_id);
CREATE INDEX idx_prendas_tipo_prenda_id ON prendas(tipo_prenda_id);
