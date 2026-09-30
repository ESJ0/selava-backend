CREATE INDEX idx_mov_inv_tipo_fecha
    ON movimientos_inventario(tipo_movimiento, fecha_movimiento);
