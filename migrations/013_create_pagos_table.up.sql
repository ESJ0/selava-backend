CREATE TABLE pagos (
    id              SERIAL PRIMARY KEY,
    pedido_id       INT           NOT NULL REFERENCES pedidos(id) ON DELETE RESTRICT,
    metodo_pago_id  INT           NOT NULL REFERENCES metodos_pago(id) ON DELETE RESTRICT,
    usuario_id      INT           NOT NULL REFERENCES usuarios(id) ON DELETE RESTRICT,
    monto           NUMERIC(10,2) NOT NULL CHECK (monto > 0),
    referencia      VARCHAR(100),
    fecha_pago      TIMESTAMP     NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pagos_pedido_id ON pagos(pedido_id);
CREATE INDEX idx_pagos_metodo_pago_id ON pagos(metodo_pago_id);
CREATE INDEX idx_pagos_usuario_id ON pagos(usuario_id);
