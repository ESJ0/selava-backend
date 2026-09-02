INSERT INTO estados_pedido (nombre, orden)
VALUES ('Cancelado', 99)
ON CONFLICT (nombre) DO NOTHING;
