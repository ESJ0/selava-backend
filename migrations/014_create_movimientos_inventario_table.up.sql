CREATE TABLE movimientos_inventario (
    id               SERIAL PRIMARY KEY,
    insumo_id        INT           NOT NULL REFERENCES insumos(id) ON DELETE RESTRICT,
    usuario_id       INT           NOT NULL REFERENCES usuarios(id) ON DELETE RESTRICT,
    tipo_movimiento  VARCHAR(10)   NOT NULL CHECK (tipo_movimiento IN ('entrada', 'salida')),
    cantidad         NUMERIC(10,2) NOT NULL CHECK (cantidad > 0),
    motivo           VARCHAR(255),
    fecha_movimiento TIMESTAMP     NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mov_inv_insumo_id ON movimientos_inventario(insumo_id);
CREATE INDEX idx_mov_inv_usuario_id ON movimientos_inventario(usuario_id);
