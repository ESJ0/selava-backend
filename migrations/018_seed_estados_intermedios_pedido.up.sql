-- Estados intermedios del flujo normal de un pedido. Junto con "Recibido"
-- (orden 1, migracion 015) y "Cancelado" (orden 99, migracion 017, fuera
-- del flujo normal), el catalogo completo queda:
-- Recibido(1) -> Rackeado(2) -> Entregado(3), con Cancelado(99) como
-- salida excepcional valida unicamente desde "Recibido" (ver Cancelar en
-- pedidoRepository.go).
INSERT INTO estados_pedido (nombre, orden) VALUES ('Rackeado', 2)  ON CONFLICT (nombre) DO NOTHING;
INSERT INTO estados_pedido (nombre, orden) VALUES ('Entregado', 3) ON CONFLICT (nombre) DO NOTHING;
