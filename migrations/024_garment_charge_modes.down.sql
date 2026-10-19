DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM prendas WHERE cantidad <> TRUNC(cantidad)) THEN
        RAISE EXCEPTION 'no se puede revertir: existen prendas con medidas decimales';
    END IF;
END $$;

ALTER TABLE prendas DROP CONSTRAINT chk_prenda_medida_modalidad;
ALTER TABLE prendas DROP CONSTRAINT chk_prenda_medida_precision;
ALTER TABLE prendas DROP CONSTRAINT chk_prenda_modalidad_cobro_aplicada;
ALTER TABLE prendas DROP COLUMN modalidad_cobro_aplicada;
ALTER TABLE prendas ALTER COLUMN cantidad TYPE INT USING cantidad::INT;

ALTER TABLE tipos_prenda DROP CONSTRAINT chk_tipo_prenda_modalidad_cobro;
ALTER TABLE tipos_prenda DROP COLUMN modalidad_cobro;
