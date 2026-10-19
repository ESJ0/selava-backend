-- No existe una correspondencia univoca entre servicios y tipos de prenda.
-- Los tipos existentes quedan pendientes de configurar, sin inventar tarifas.
-- NUMERIC sin escala evita redondear silenciosamente entradas con >2 decimales.
ALTER TABLE tipos_prenda ADD COLUMN precio NUMERIC;
ALTER TABLE tipos_prenda ADD CONSTRAINT chk_tipo_prenda_precio
    CHECK (precio > 0 AND precio <= 99999999.99 AND precio = ROUND(precio, 2));

CREATE FUNCTION exigir_precio_tipo_prenda() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.precio IS NULL AND (TG_OP = 'INSERT' OR OLD.precio IS NOT NULL) THEN
        RAISE EXCEPTION 'precio es requerido' USING ERRCODE = '23502';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER tipo_prenda_precio_requerido BEFORE INSERT OR UPDATE ON tipos_prenda
    FOR EACH ROW EXECUTE FUNCTION exigir_precio_tipo_prenda();

-- NULL identifica exclusivamente las prendas anteriores a SEL-107.
-- Su calculo sigue usando los importes congelados de prenda_servicios.
ALTER TABLE prendas ADD COLUMN precio_aplicado NUMERIC;
ALTER TABLE prendas ADD CONSTRAINT chk_prenda_precio_aplicado
    CHECK (precio_aplicado > 0 AND precio_aplicado <= 99999999.99
           AND precio_aplicado = ROUND(precio_aplicado, 2));

-- Se conserva la columna y sus valores para consumidores anteriores.
-- Los servicios nuevos pueden omitirla y no interviene en nuevos cobros.
ALTER TABLE servicios ALTER COLUMN precio_base DROP NOT NULL;

COMMENT ON COLUMN tipos_prenda.precio IS 'Tarifa vigente. NULL solo durante la configuracion de tipos anteriores a SEL-107.';
COMMENT ON COLUMN prendas.precio_aplicado IS 'Tarifa del tipo congelada al registrar la prenda. NULL identifica calculo historico por servicios.';
COMMENT ON COLUMN servicios.precio_base IS 'Deprecado: compatibilidad historica; no determina el cobro de prendas nuevas.';
