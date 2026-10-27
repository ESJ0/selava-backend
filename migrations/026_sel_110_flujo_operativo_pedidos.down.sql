DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pedidos p
        JOIN estados_pedido e ON e.id = p.estado_actual_id
        WHERE e.codigo IN ('EN_LAVADO', 'EN_LAVADO_SECO', 'EN_PLANCHADO')
    ) OR EXISTS (
        SELECT 1
        FROM pedido_estados_historial h
        JOIN estados_pedido e ON e.id = h.estado_id
        WHERE e.codigo IN ('EN_LAVADO', 'EN_LAVADO_SECO', 'EN_PLANCHADO')
    ) THEN
        RAISE EXCEPTION 'no se puede revertir SEL-110: existen pedidos en las nuevas etapas operativas';
    END IF;
END $$;

DROP TRIGGER trg_sincronizar_etapa_servicio ON servicios;
DROP FUNCTION sincronizar_etapa_servicio();
DROP TABLE estado_pedido_servicios;

DELETE FROM estados_pedido
WHERE codigo IN ('EN_LAVADO', 'EN_LAVADO_SECO', 'EN_PLANCHADO');

UPDATE estados_pedido SET nombre = 'Entregado' WHERE codigo = 'FINALIZADO';
UPDATE estados_pedido SET orden = 1 WHERE codigo = 'RECIBIDO';
UPDATE estados_pedido SET orden = 2 WHERE codigo = 'RACKEADO';
UPDATE estados_pedido SET orden = 3 WHERE codigo = 'FINALIZADO';
UPDATE estados_pedido SET orden = 99 WHERE codigo = 'CANCELADO';

ALTER TABLE estados_pedido
    DROP COLUMN es_cancelacion,
    DROP COLUMN es_terminal,
    DROP COLUMN es_obligatorio,
    DROP COLUMN codigo;

