-- SEL-110: flujo operativo secuencial y etapas condicionales por servicio.
-- Se conservan los IDs existentes para no romper pedidos ni historiales.

ALTER TABLE estados_pedido
    ADD COLUMN codigo VARCHAR(50),
    ADD COLUMN es_obligatorio BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN es_terminal BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN es_cancelacion BOOLEAN NOT NULL DEFAULT FALSE;

-- El estado terminal anterior pasa a llamarse como lo define SEL-110.
UPDATE estados_pedido SET nombre = 'Finalizado' WHERE nombre = 'Entregado';

UPDATE estados_pedido
SET codigo = CASE nombre
    WHEN 'Recibido' THEN 'RECIBIDO'
    WHEN 'Rackeado' THEN 'RACKEADO'
    WHEN 'Finalizado' THEN 'FINALIZADO'
    WHEN 'Cancelado' THEN 'CANCELADO'
    ELSE 'LEGADO_' || id
END;

ALTER TABLE estados_pedido
    ALTER COLUMN codigo SET NOT NULL,
    ADD CONSTRAINT estados_pedido_codigo_key UNIQUE (codigo);

-- Las decenas dejan espacio para insertar nuevas etapas sin reordenar el flujo.
UPDATE estados_pedido SET orden = 10, es_obligatorio = TRUE WHERE codigo = 'RECIBIDO';
UPDATE estados_pedido SET orden = 50, es_obligatorio = TRUE WHERE codigo = 'RACKEADO';
UPDATE estados_pedido SET orden = 60, es_obligatorio = TRUE, es_terminal = TRUE WHERE codigo = 'FINALIZADO';
UPDATE estados_pedido SET orden = 99, es_cancelacion = TRUE, es_terminal = TRUE WHERE codigo = 'CANCELADO';

INSERT INTO estados_pedido (nombre, codigo, orden)
VALUES
    ('En lavado', 'EN_LAVADO', 20),
    ('En lavado en seco', 'EN_LAVADO_SECO', 30),
    ('En planchado', 'EN_PLANCHADO', 40)
ON CONFLICT (codigo) DO NOTHING;

CREATE TABLE estado_pedido_servicios (
    estado_pedido_id INT NOT NULL REFERENCES estados_pedido(id) ON DELETE CASCADE,
    servicio_id      INT NOT NULL REFERENCES servicios(id) ON DELETE CASCADE,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (estado_pedido_id, servicio_id)
);

CREATE INDEX idx_estado_pedido_servicios_servicio
    ON estado_pedido_servicios(servicio_id);

-- Relación inicial del catálogo auditado. Un servicio puede habilitar una etapa
-- y una etapa puede aplicar a varios servicios sin depender de sus IDs.
INSERT INTO estado_pedido_servicios (estado_pedido_id, servicio_id)
SELECT e.id, s.id
FROM estados_pedido e
JOIN servicios s ON
    (e.codigo = 'EN_LAVADO' AND LOWER(s.nombre) = 'lavado') OR
    (e.codigo = 'EN_LAVADO_SECO' AND LOWER(s.nombre) = 'lavado en seco') OR
    (e.codigo = 'EN_PLANCHADO' AND LOWER(s.nombre) = 'planchado')
ON CONFLICT DO NOTHING;

-- Mantiene la relación en instalaciones donde el seed se ejecuta después de
-- migrar y cuando el nombre de un servicio reconocido cambia.
CREATE OR REPLACE FUNCTION sincronizar_etapa_servicio() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM estado_pedido_servicios eps
    USING estados_pedido e
    WHERE eps.estado_pedido_id = e.id
      AND eps.servicio_id = NEW.id
      AND e.codigo IN ('EN_LAVADO', 'EN_LAVADO_SECO', 'EN_PLANCHADO');

    INSERT INTO estado_pedido_servicios (estado_pedido_id, servicio_id)
    SELECT e.id, NEW.id
    FROM estados_pedido e
    WHERE e.codigo = CASE LOWER(NEW.nombre)
        WHEN 'lavado' THEN 'EN_LAVADO'
        WHEN 'lavado en seco' THEN 'EN_LAVADO_SECO'
        WHEN 'planchado' THEN 'EN_PLANCHADO'
    END
    ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sincronizar_etapa_servicio
AFTER INSERT OR UPDATE OF nombre ON servicios
FOR EACH ROW EXECUTE FUNCTION sincronizar_etapa_servicio();

