CREATE INDEX idx_pedidos_fecha_recibido_id ON pedidos(fecha_recibido DESC, id DESC);
CREATE INDEX idx_pedidos_estado_fecha_recibido_id ON pedidos(estado_actual_id, fecha_recibido DESC, id DESC);
