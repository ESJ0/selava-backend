CREATE TABLE pedido_estados_historial (
    id            SERIAL PRIMARY KEY,
    pedido_id     INT         NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    estado_id     INT         NOT NULL REFERENCES estados_pedido(id) ON DELETE RESTRICT,
    usuario_id    INT         NOT NULL REFERENCES usuarios(id) ON DELETE RESTRICT,
    fecha_cambio  TIMESTAMP   NOT NULL DEFAULT NOW(),
    observaciones VARCHAR(255),
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_peh_pedido_id ON pedido_estados_historial(pedido_id);
CREATE INDEX idx_peh_estado_id ON pedido_estados_historial(estado_id);
CREATE INDEX idx_peh_usuario_id ON pedido_estados_historial(usuario_id);
