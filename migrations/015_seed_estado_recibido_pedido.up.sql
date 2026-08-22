INSERT INTO estados_pedido (nombre, orden)
VALUES ('Recibido', 1)
ON CONFLICT (nombre) DO NOTHING;
