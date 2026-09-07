CREATE TABLE pedidos (
    id                     SERIAL PRIMARY KEY,
    cliente_id             INT           NOT NULL REFERENCES clientes(id) ON DELETE RESTRICT,
    usuario_id             INT           NOT NULL REFERENCES usuarios(id) ON DELETE RESTRICT,
    estado_actual_id       INT           NOT NULL REFERENCES estados_pedido(id) ON DELETE RESTRICT,
    fecha_recibido         TIMESTAMP     NOT NULL DEFAULT NOW(),
    fecha_entrega_estimada TIMESTAMP,
    fecha_entrega_real     TIMESTAMP,
    total                  NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    observaciones          TEXT,
    activo                 BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at             TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pedidos_cliente_id ON pedidos(cliente_id);
CREATE INDEX idx_pedidos_usuario_id ON pedidos(usuario_id);
CREATE INDEX idx_pedidos_estado_actual_id ON pedidos(estado_actual_id);
