CREATE TABLE notificaciones (
    id         SERIAL PRIMARY KEY,
    cliente_id INT       NOT NULL REFERENCES clientes(id) ON DELETE CASCADE,
    pedido_id  INT       REFERENCES pedidos(id) ON DELETE SET NULL,
    mensaje    TEXT      NOT NULL CHECK (LENGTH(TRIM(mensaje)) > 0),
    leida      BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notificaciones_cliente_created_at
    ON notificaciones(cliente_id, created_at DESC, id DESC);

CREATE INDEX idx_notificaciones_cliente_no_leidas
    ON notificaciones(cliente_id, created_at DESC)
    WHERE leida = FALSE;
